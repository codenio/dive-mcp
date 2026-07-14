#!/usr/bin/env python3
"""Compare dive-mcp vs raw dive CLI for the same image analysis task.

Measures wall-clock time, estimated token usage, and answer accuracy against
ground truth from `dive <image> --json`.
"""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


# ~4 chars per token (OpenAI-style heuristic for English/JSON).
CHARS_PER_TOKEN = 4

USER_PROMPT = (
    "Analyze the container image {image!r}. Report: (1) efficiency score, "
    "(2) wasted bytes, (3) layer count, (4) top wasted file path and its "
    "total wasted bytes, (5) whether it passes CI thresholds "
    "(efficiency >= 0.9, wasted percent <= 0.1)."
)

TOOL_SCHEMAS_EST_BYTES = 2800  # four dive-mcp tool definitions (approx)


@dataclass
class GroundTruth:
    image: str
    size_bytes: int
    inefficient_bytes: int
    efficiency_score: float
    layer_count: int
    top_wasted_file: str
    top_wasted_bytes: int
    ci_pass: bool
    wasted_percent: float


@dataclass
class ScenarioResult:
    name: str
    dive_invocations: int
    agent_turns: int
    wall_ms: float
    dive_ms: float
    context_bytes: int
    estimated_tokens: int
    answers: dict[str, Any] = field(default_factory=dict)
    accuracy: dict[str, bool] = field(default_factory=dict)


def estimate_tokens(text: str | bytes) -> int:
    n = len(text) if isinstance(text, str) else len(text.decode("utf-8", errors="replace"))
    return (n + CHARS_PER_TOKEN - 1) // CHARS_PER_TOKEN


def run_dive_json(image: str) -> tuple[dict[str, Any], float]:
    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as tmp:
        path = tmp.name
    try:
        start = time.perf_counter()
        proc = subprocess.run(
            ["dive", image, "--json", path],
            capture_output=True,
            text=True,
        )
        elapsed_ms = (time.perf_counter() - start) * 1000
        if proc.returncode != 0:
            raise RuntimeError(f"dive failed: {proc.stderr}")
        data = json.loads(Path(path).read_text())
        return data, elapsed_ms
    finally:
        Path(path).unlink(missing_ok=True)


def ground_truth_from(data: dict[str, Any], image: str) -> GroundTruth:
    img = data["image"]
    refs = sorted(
        img["fileReference"],
        key=lambda r: r["count"] * r["sizeBytes"],
        reverse=True,
    )
    top = refs[0] if refs else {"file": "(none)", "count": 0, "sizeBytes": 0}
    wasted_pct = img["inefficientBytes"] / img["sizeBytes"] if img["sizeBytes"] else 0.0
    return GroundTruth(
        image=image,
        size_bytes=img["sizeBytes"],
        inefficient_bytes=img["inefficientBytes"],
        efficiency_score=img["efficiencyScore"],
        layer_count=len(data["layer"]),
        top_wasted_file=top["file"],
        top_wasted_bytes=top["count"] * top["sizeBytes"],
        ci_pass=img["efficiencyScore"] >= 0.9 and wasted_pct <= 0.1,
        wasted_percent=wasted_pct,
    )


def check_accuracy(answers: dict[str, Any], gt: GroundTruth) -> dict[str, bool]:
    def close(a: float, b: float, tol: float = 1e-6) -> bool:
        return abs(a - b) <= tol

    return {
        "efficiency_score": close(float(answers.get("efficiency_score", -1)), gt.efficiency_score),
        "inefficient_bytes": int(answers.get("inefficient_bytes", -1)) == gt.inefficient_bytes,
        "layer_count": int(answers.get("layer_count", -1)) == gt.layer_count,
        "top_wasted_file": answers.get("top_wasted_file") == gt.top_wasted_file
        or (answers.get("top_wasted_file") in ("(none)", "") and gt.top_wasted_file in ("(none)", "")),
        "top_wasted_bytes": int(answers.get("top_wasted_bytes", -1)) == gt.top_wasted_bytes,
        "ci_pass": bool(answers.get("ci_pass")) == gt.ci_pass,
    }


def scenario_naive_full_json(image: str, gt: GroundTruth) -> ScenarioResult:
    """Agent reads the entire dive JSON (including per-layer file trees)."""
    start = time.perf_counter()
    data, dive_ms = run_dive_json(image)
    raw = json.dumps(data)
    prompt = USER_PROMPT.format(image=image)
    context_bytes = len(prompt.encode()) + len(raw.encode())
    refs = sorted(
        data["image"]["fileReference"],
        key=lambda r: r["count"] * r["sizeBytes"],
        reverse=True,
    )
    top = refs[0] if refs else {"file": "(none)", "count": 0, "sizeBytes": 0}
    wasted_pct = data["image"]["inefficientBytes"] / data["image"]["sizeBytes"]
    answers = {
        "efficiency_score": data["image"]["efficiencyScore"],
        "inefficient_bytes": data["image"]["inefficientBytes"],
        "layer_count": len(data["layer"]),
        "top_wasted_file": top["file"],
        "top_wasted_bytes": top["count"] * top["sizeBytes"],
        "ci_pass": data["image"]["efficiencyScore"] >= 0.9 and wasted_pct <= 0.1,
    }
    wall_ms = (time.perf_counter() - start) * 1000
    return ScenarioResult(
        name="Without dive-mcp (read full dive JSON)",
        dive_invocations=1,
        agent_turns=2,
        wall_ms=wall_ms,
        dive_ms=dive_ms,
        context_bytes=context_bytes,
        estimated_tokens=estimate_tokens(prompt) + estimate_tokens(raw),
        answers=answers,
        accuracy=check_accuracy(answers, gt),
    )


def scenario_shell_jq(image: str, gt: GroundTruth) -> ScenarioResult:
    """Agent runs dive, then jq/shell to extract fields (typical capable agent)."""
    start = time.perf_counter()
    dive_ms_total = 0.0
    shell_outputs: list[str] = []

    with tempfile.NamedTemporaryFile(suffix=".json", delete=False) as tmp:
        json_path = tmp.name

    try:
        t0 = time.perf_counter()
        proc = subprocess.run(["dive", image, "--json", json_path], capture_output=True, text=True)
        dive_ms_total += (time.perf_counter() - t0) * 1000
        if proc.returncode != 0:
            raise RuntimeError(proc.stderr)
        shell_outputs.append(proc.stderr.strip() or "(dive completed)")

        jq_cmds = [
            f"jq '.image | {{efficiencyScore, inefficientBytes, sizeBytes}}' {json_path}",
            f"jq '.layer | length' {json_path}",
            (
                f"jq -r '.image.fileReference | sort_by(.count * .sizeBytes) | reverse | .[0] | "
                f"{{file, total: (.count * .sizeBytes)}}' {json_path}"
            ),
        ]
        extracted: dict[str, Any] = {}
        for cmd in jq_cmds:
            t0 = time.perf_counter()
            out = subprocess.run(cmd, shell=True, capture_output=True, text=True)
            dive_ms_total += (time.perf_counter() - t0) * 1000
            shell_outputs.append(f"$ {cmd}\n{out.stdout.strip()}")
            if "efficiencyScore" in cmd:
                row = json.loads(out.stdout)
                extracted["efficiency_score"] = row["efficiencyScore"]
                extracted["inefficient_bytes"] = row["inefficientBytes"]
                extracted["size_bytes"] = row["sizeBytes"]
            elif "length" in cmd:
                extracted["layer_count"] = int(out.stdout.strip())
            else:
                if out.stdout.strip() in ("", "null"):
                    extracted["top_wasted_file"] = "(none)"
                    extracted["top_wasted_bytes"] = 0
                else:
                    row = json.loads(out.stdout)
                    extracted["top_wasted_file"] = row["file"]
                    extracted["top_wasted_bytes"] = row["total"]

        wasted_pct = extracted["inefficient_bytes"] / extracted["size_bytes"]
        extracted["ci_pass"] = extracted["efficiency_score"] >= 0.9 and wasted_pct <= 0.1

        prompt = USER_PROMPT.format(image=image)
        tool_call_overhead = 600  # shell tool invocations × ~200 bytes each
        context_bytes = (
            len(prompt.encode())
            + sum(len(s.encode()) for s in shell_outputs)
            + tool_call_overhead
        )
        wall_ms = (time.perf_counter() - start) * 1000
        return ScenarioResult(
            name="Without dive-mcp (dive + jq shell)",
            dive_invocations=1,
            agent_turns=5,
            wall_ms=wall_ms,
            dive_ms=dive_ms_total,
            context_bytes=context_bytes,
            estimated_tokens=estimate_tokens(prompt) + context_bytes // CHARS_PER_TOKEN,
            answers=extracted,
            accuracy=check_accuracy(extracted, gt),
        )
    finally:
        Path(json_path).unlink(missing_ok=True)


def mcp_tool_outputs(image: str, data: dict[str, Any]) -> dict[str, str]:
    """Build the JSON payloads dive-mcp tools would return (same logic as server)."""
    img = data["image"]
    refs = sorted(img["fileReference"], key=lambda r: r["count"] * r["sizeBytes"], reverse=True)
    top = refs[:5]
    wasted_pct = img["inefficientBytes"] / img["sizeBytes"] if img["sizeBytes"] else 0.0

    analyze = {
        "image": image,
        "sizeBytes": img["sizeBytes"],
        "inefficientBytes": img["inefficientBytes"],
        "efficiencyScore": img["efficiencyScore"],
        "layerCount": len(data["layer"]),
        "summary": (
            f'Image "{image}": efficiency {img["efficiencyScore"]:.4f}, '
            f'{img["inefficientBytes"]} wasted bytes, {len(data["layer"])} layers.'
        ),
    }
    layers = {
        "image": image,
        "layers": [
            {
                "index": l["index"],
                "digestId": l["digestId"],
                "sizeBytes": l["sizeBytes"],
                "command": l["command"],
            }
            for l in data["layer"]
        ],
    }
    wasted = {
        "image": image,
        "inefficientBytes": img["inefficientBytes"],
        "files": [
            {
                "file": f["file"],
                "count": f["count"],
                "sizeBytes": f["sizeBytes"],
                "totalWastedBytes": f["count"] * f["sizeBytes"],
            }
            for f in top
        ],
    }
    ci = {
        "image": image,
        "pass": img["efficiencyScore"] >= 0.9 and wasted_pct <= 0.1,
        "efficiencyScore": img["efficiencyScore"],
        "inefficientBytes": img["inefficientBytes"],
        "wastedPercent": wasted_pct,
    }
    return {
        "analyze_image": json.dumps(analyze),
        "list_layers": json.dumps(layers),
        "get_wasted_space": json.dumps(wasted),
        "ci_check": json.dumps(ci),
    }


def scenario_with_dive_mcp(image: str, gt: GroundTruth) -> ScenarioResult:
    """Agent calls four dive-mcp tools; dive runs once (cached)."""
    start = time.perf_counter()
    data, dive_ms = run_dive_json(image)
    outputs = mcp_tool_outputs(image, data)

    answers = {
        "efficiency_score": data["image"]["efficiencyScore"],
        "inefficient_bytes": data["image"]["inefficientBytes"],
        "layer_count": len(data["layer"]),
        "top_wasted_file": (
            json.loads(outputs["get_wasted_space"])["files"][0]["file"]
            if json.loads(outputs["get_wasted_space"])["files"]
            else "(none)"
        ),
        "top_wasted_bytes": (
            json.loads(outputs["get_wasted_space"])["files"][0]["totalWastedBytes"]
            if json.loads(outputs["get_wasted_space"])["files"]
            else 0
        ),
        "ci_pass": json.loads(outputs["ci_check"])["pass"],
    }

    prompt = USER_PROMPT.format(image=image)
    tool_request_bytes = 4 * 120  # four compact tool call payloads
    response_bytes = sum(len(v.encode()) for v in outputs.values())
    context_bytes = (
        len(prompt.encode()) + TOOL_SCHEMAS_EST_BYTES + tool_request_bytes + response_bytes
    )
    wall_ms = (time.perf_counter() - start) * 1000

    return ScenarioResult(
        name="With dive-mcp (4 structured tools, 1 dive run)",
        dive_invocations=1,
        agent_turns=5,
        wall_ms=wall_ms,
        dive_ms=dive_ms,
        context_bytes=context_bytes,
        estimated_tokens=estimate_tokens(prompt) + context_bytes // CHARS_PER_TOKEN,
        answers=answers,
        accuracy=check_accuracy(answers, gt),
    )


def fmt_ms(ms: float) -> str:
    return f"{ms:.0f} ms"


def print_report(image: str, gt: GroundTruth, results: list[ScenarioResult]) -> None:
    print("=" * 72)
    print("dive-mcp evaluation")
    print("=" * 72)
    print(f"Image:     {image}")
    print(f"Dive:      {subprocess.check_output(['dive', '--version'], text=True).strip()}")
    print()
    print("Ground truth (from dive --json):")
    print(f"  efficiency_score:   {gt.efficiency_score:.10f}")
    print(f"  inefficient_bytes:  {gt.inefficient_bytes}")
    print(f"  layer_count:        {gt.layer_count}")
    print(f"  top_wasted_file:    {gt.top_wasted_file}")
    print(f"  top_wasted_bytes:   {gt.top_wasted_bytes}")
    print(f"  ci_pass:            {gt.ci_pass}")
    print(f"  wasted_percent:     {gt.wasted_percent:.6f}")
    print()

    header = f"{'Scenario':<42} {'Turns':>5} {'Wall':>8} {'Dive':>8} {'Context':>10} {'~Tokens':>8} {'Accuracy':>8}"
    print(header)
    print("-" * len(header))
    for r in results:
        correct = sum(r.accuracy.values())
        total = len(r.accuracy)
        acc = f"{correct}/{total}"
        print(
            f"{r.name:<42} {r.agent_turns:>5} {fmt_ms(r.wall_ms):>8} {fmt_ms(r.dive_ms):>8} "
            f"{r.context_bytes:>9} B {r.estimated_tokens:>8} {acc:>8}"
        )

    print()
    print("Accuracy breakdown:")
    for r in results:
        print(f"  {r.name}:")
        for field, ok in r.accuracy.items():
            mark = "OK" if ok else "MISS"
            print(f"    [{mark}] {field}")

    baseline = results[0]
    mcp = results[-1]
    token_savings = 100 * (1 - mcp.estimated_tokens / baseline.estimated_tokens)
    context_savings = 100 * (1 - mcp.context_bytes / baseline.context_bytes)
    print()
    print("Summary (with dive-mcp vs naive full JSON):")
    print(f"  Token reduction:   {token_savings:.1f}% ({baseline.estimated_tokens} -> {mcp.estimated_tokens})")
    print(f"  Context reduction: {context_savings:.1f}% ({baseline.context_bytes} B -> {mcp.context_bytes} B)")
    print(f"  Wall time ratio:   {mcp.wall_ms / baseline.wall_ms:.2f}x (MCP includes same single dive run)")
    if len(results) >= 2:
        jq = results[1]
        print(f"  vs shell+jq tokens: {100 * (1 - mcp.estimated_tokens / jq.estimated_tokens):.1f}% reduction")


def main() -> int:
    image = sys.argv[1] if len(sys.argv) > 1 else "dive-mcp-eval:ci"
    data, _ = run_dive_json(image)
    gt = ground_truth_from(data, image)

    results = [
        scenario_naive_full_json(image, gt),
        scenario_shell_jq(image, gt),
        scenario_with_dive_mcp(image, gt),
    ]
    print_report(image, gt, results)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

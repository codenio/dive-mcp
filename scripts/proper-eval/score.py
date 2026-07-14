#!/usr/bin/env python3
"""Score blind agent evaluation results against ground truth."""

from __future__ import annotations

import json
import sys
from pathlib import Path

FIELDS = [
    "efficiency_score",
    "inefficient_bytes",
    "layer_count",
    "top_wasted_file",
    "top_wasted_bytes",
    "ci_pass",
]


def score(agent: dict, gt: dict) -> tuple[int, dict[str, bool]]:
    answers = agent["answers"]
    checks: dict[str, bool] = {}
    for f in FIELDS:
        if f == "efficiency_score":
            checks[f] = abs(float(answers[f]) - float(gt[f])) < 1e-9
        else:
            checks[f] = answers[f] == gt[f]
    return sum(checks.values()), checks


def main() -> int:
    path = Path(sys.argv[1] if len(sys.argv) > 1 else Path(__file__).with_name("results-blind-2026-07-11.json"))
    data = json.loads(path.read_text())
    gt = data["ground_truth"]

    for key in ("agent_a_shell_only", "agent_b_dive_mcp"):
        agent = data[key]
        correct, checks = score(agent, gt)
        print(f"\n{key}: {correct}/{len(FIELDS)}")
        for field, ok in checks.items():
            print(f"  [{'OK' if ok else 'MISS'}] {field}: {agent['answers'][field]} (expected {gt[field]})")
        m = agent["metrics"]
        io = m["tool_input_chars"] + m["tool_output_chars"]
        print(f"  wall={m['wall_time_seconds']}s tools={m['tool_call_count']} io_chars={io} (~{io//4} tokens)")

    cmp = data["comparison"]
    print("\nComparison (B vs A tool I/O):")
    print(f"  tool I/O chars: {cmp['tool_io_chars_agent_a']} -> {cmp['tool_io_chars_agent_b']} ({cmp['tool_io_reduction_percent']}% reduction)")
    print(f"  wall time: {data['agent_a_shell_only']['metrics']['wall_time_seconds']}s -> {data['agent_b_dive_mcp']['metrics']['wall_time_seconds']}s")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

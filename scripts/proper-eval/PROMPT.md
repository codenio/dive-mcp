# Blind agent evaluation prompt

Use this **exact** user message for both agents (replace `{IMAGE}`):

```
Analyze the container image `{IMAGE}`.

Return your answer as a single JSON code block with this exact schema (no extra keys):

{
  "efficiency_score": <float>,
  "inefficient_bytes": <integer>,
  "layer_count": <integer>,
  "top_wasted_file": <string>,
  "top_wasted_bytes": <integer>,
  "ci_pass": <boolean>,
  "metrics": {
    "wall_time_seconds": <float>,
    "tool_call_count": <integer>,
    "tool_input_chars": <integer>,
    "tool_output_chars": <integer>,
    "approach": "<one sentence describing steps taken>"
  }
}

Rules for the analysis fields:
- CI pass means efficiency >= 0.9 AND (inefficient_bytes / total_size_bytes) <= 0.1
- top_wasted_file: use "(none)" if no wasted files
- top_wasted_bytes: count * sizeBytes for the worst offender

Rules for metrics:
- wall_time_seconds: time from receiving this message until JSON is ready
- Count every tool invocation you make
- tool_input_chars / tool_output_chars: sum of character lengths of your tool arguments and tool results
```

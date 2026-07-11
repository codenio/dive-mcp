# Architecture

`dive-mcp` is a thin stdio MCP server. It does not embed dive's analysis logic — it
shells out to the `dive` CLI, parses JSON output, and exposes the result as MCP tools.

## Invocation

```sh
dive <image> --json <tmpfile> [--source docker|podman|docker-archive]
```

The temp file is always removed after parsing, even on error.

## JSON output

`dive` writes a document shaped like:

```json
{
  "layer": [
    {"index": 0, "id": "blobs", "digestId": "sha256:...", "sizeBytes": 8652792, "command": "ADD ... # buildkit"}
  ],
  "image": {
    "sizeBytes": 8755204,
    "inefficientBytes": 102412,
    "efficiencyScore": 0.9883034136,
    "fileReference": [
      {"count": 2, "sizeBytes": 102400, "file": "/tmp/big.bin"}
    ]
  }
}
```

Go types for this structure live in [`internal/dive/types.go`](../internal/dive/types.go).

## CI threshold logic

`ci_check` implements dive's `--ci` threshold semantics in Go rather than shelling out to
`dive --ci`, so the MCP tool can return structured pass/fail data instead of a bare
process exit code.

## Caching

The [`Runner`](../internal/dive/dive.go) caches parsed analyses for the process lifetime,
keyed by `(image, source)`. Multiple tool calls against the same image in one MCP session
invoke `dive` only once.

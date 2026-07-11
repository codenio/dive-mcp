# MCP tools

All four tools share a per-process cache keyed by `(image, source)`, so calling multiple
tools against the same image within one MCP session only invokes the `dive` CLI once.

| Tool | Input | Output |
|---|---|---|
| `analyze_image` | `image` (required), `source` (optional: `docker`\|`podman`\|`docker-archive`) | Total image size, efficiency score (0-1), inefficient (wasted) bytes, layer count, human-readable summary. |
| `list_layers` | `image`, `source` (optional) | Per-layer `index`, `digestId`, `sizeBytes`, `command`. |
| `get_wasted_space` | `image`, `source` (optional), `limit` (optional, default 20) | Top files duplicated across layers, sorted descending by `count * sizeBytes`. |
| `ci_check` | `image`, `source` (optional), `lowest_efficiency` (default 0.9), `highest_wasted_bytes` (default disabled), `highest_user_wasted_percent` (default 0.1) | `pass`/`fail` plus human-readable `reasons`, mirroring `dive --ci` threshold semantics. |

## `source` values

- `docker` — default; uses the Docker engine
- `podman` — uses Podman
- `docker-archive` — inspect a `docker save` tarball (no running daemon required)

## `ci_check` thresholds

Mirrors `dive --ci` semantics but returns structured JSON instead of a process exit code:

- `lowest_efficiency` — minimum allowed efficiency score (default `0.9`)
- `highest_wasted_bytes` — maximum inefficient bytes; `0` disables this check
- `highest_user_wasted_percent` — maximum wasted fraction of image size (default `0.1`)

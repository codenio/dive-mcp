# dive version compatibility

`dive-mcp` shells out to the external `dive` CLI and parses its `--json` output. It does
**not** vendor or pin a specific `dive` release at build time — users install `dive`
separately and point `dive-mcp` at it via `PATH` or `DIVE_MCP_DIVE_PATH`.

## Supported versions

| dive release | Status | Notes |
|---|---|---|
| `v0.13.x` | Supported (tested) | Current upstream; recommended |
| `v0.12.x` | Supported (tested) | Oldest version in the CI matrix |
| `v0.11.x` | Best effort | Not exercised in CI |
| `v0.10.x` | Best effort | Not exercised in CI |
| `< v0.10.0` | Best effort | `--json` existed earlier but is not exercised in CI |

## Maintaining compatibility

1. **When upstream `dive` ships a new release**, add the version to the matrix in
   [`.github/workflows/dive-compat.yml`](https://github.com/codenio/dive-mcp/blob/main/.github/workflows/dive-compat.yml) and open a
   PR. The weekly scheduled job will also catch drift if you forget.
2. **If `dive` changes its JSON shape**, update
   [`internal/dive/types.go`](https://github.com/codenio/dive-mcp/blob/main/internal/dive/types.go) and refresh
   [`internal/dive/testdata/sample.json`](https://github.com/codenio/dive-mcp/blob/main/internal/dive/testdata/sample.json) from a
   real `dive <image> --json sample.json` run.
3. **For older `dive` installs**, no special `dive-mcp` build is required — any
   `dive-mcp` release works as long as the installed `dive` produces parseable JSON.
   Pin `dive` in your environment if you need reproducibility.
4. **Local compatibility check** (requires Docker):

   ```sh
   ./scripts/install-dive.sh 0.13.1
   ./scripts/verify-dive-compat.sh ./.dive-bin/dive
   ```

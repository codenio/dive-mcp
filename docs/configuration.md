# Configuration

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `DIVE_MCP_DIVE_PATH` | (unset — resolved via `PATH`) | Explicit path to the `dive` binary. Overrides `PATH` lookup. |
| `DIVE_MCP_TIMEOUT` | `5m` | Per-invocation timeout for running `dive`, as a Go duration string (e.g. `90s`, `10m`). Falls back to the default if unset or unparsable. |

## Claude Code

```sh
claude mcp add dive-mcp -- dive-mcp
```

See [`examples/claude-code.sh`](https://github.com/codenio/dive-mcp/blob/main/examples/claude-code.sh)
for a version with explicit env vars.

## Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "dive-mcp": {
      "command": "dive-mcp",
      "args": []
    }
  }
}
```

See [`examples/claude_desktop_config.json`](https://github.com/codenio/dive-mcp/blob/main/examples/claude_desktop_config.json).

## Cursor

Add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "dive-mcp": {
      "command": "dive-mcp",
      "args": []
    }
  }
}
```

See [`examples/cursor-mcp.json`](https://github.com/codenio/dive-mcp/blob/main/examples/cursor-mcp.json).

## Codex

Add to `~/.codex/config.toml`:

```toml
[mcp_servers.dive-mcp]
command = "dive-mcp"
args = []
```

See [`examples/codex-config.toml`](https://github.com/codenio/dive-mcp/blob/main/examples/codex-config.toml).

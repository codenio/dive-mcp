# dive-mcp

[![CI](https://github.com/codenio/dive-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/codenio/dive-mcp/actions/workflows/ci.yml)
[![Dive compatibility](https://github.com/codenio/dive-mcp/actions/workflows/dive-compat.yml/badge.svg)](https://github.com/codenio/dive-mcp/actions/workflows/dive-compat.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/codenio/dive-mcp)](https://github.com/codenio/dive-mcp/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/codenio/dive-mcp/cmd/dive-mcp.svg)](https://pkg.go.dev/github.com/codenio/dive-mcp/cmd/dive-mcp)
[![MCP](https://img.shields.io/badge/MCP-compatible-6366F1)](https://modelcontextprotocol.io)
[![dive](https://img.shields.io/badge/dive-v0.12%2B-blue)](https://github.com/wagoodman/dive)

MCP server for [`dive`](https://github.com/wagoodman/dive) — inspect container image
layers, find wasted space, and run CI-style efficiency checks from AI coding assistants.

**dive compatibility:** [`dive`](https://github.com/wagoodman/dive) v0.12.0 – v0.13.x (CI-tested).
Uses your installed `dive` binary — not bundled. [Details →](docs/dive-compatibility.md)

## Quick start

**1. Install prerequisites**

```sh
# dive — https://github.com/wagoodman/dive
brew install dive          # macOS
# or: https://github.com/wagoodman/dive#installation

# Docker or Podman must be running for image analysis
docker info >/dev/null
```

**2. Install dive-mcp**

```sh
go install github.com/codenio/dive-mcp/cmd/dive-mcp@latest
dive-mcp --version
```

**3. Add to your MCP client**

**Cursor** — `.cursor/mcp.json`:

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

**Claude Code**:

```sh
claude mcp add dive-mcp -- dive-mcp
```

**Codex** — `~/.codex/config.toml`:

```toml
[mcp_servers.dive-mcp]
command = "dive-mcp"
args = []
```

Claude Desktop and env vars: [docs/configuration.md](docs/configuration.md) · [`examples/`](examples/)

**4. Analyze an image**

Restart your MCP client, then ask your assistant:

> Analyze `alpine:latest` with dive-mcp and tell me the efficiency score and top wasted files.

The assistant can call `analyze_image`, `list_layers`, `get_wasted_space`, and `ci_check`.
See [docs/tools.md](docs/tools.md) for the full tool reference.

## Tools

- `analyze_image` — image size, efficiency score, wasted bytes, layer count
- `list_layers` — per-layer digest, size, and build command
- `get_wasted_space` — top duplicated files across layers
- `ci_check` — pass/fail against efficiency and waste thresholds

See [docs/tools.md](docs/tools.md) for parameters and output fields.

Build from source: `git clone` → `make build` — see [docs/development.md](docs/development.md).

## Documentation

| Doc | Contents |
|---|---|
| [configuration.md](docs/configuration.md) | Client setup and environment variables |
| [tools.md](docs/tools.md) | MCP tool reference |
| [development.md](docs/development.md) | Build, test, pre-commit hooks |
| [architecture.md](docs/architecture.md) | How dive-mcp invokes and parses dive |
| [dive-compatibility.md](docs/dive-compatibility.md) | Supported dive versions |
| [releasing.md](docs/releasing.md) | Maintainer release workflow |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).

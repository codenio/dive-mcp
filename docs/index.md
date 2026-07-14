<div class="dm-hero" markdown>

# dive-mcp

MCP server for [dive](https://github.com/wagoodman/dive) — inspect container image
layers, find wasted space, and run CI-style efficiency checks from AI coding assistants.

<div class="dm-badges" markdown>

[ :octicons-mark-github-16: GitHub ](https://github.com/codenio/dive-mcp){ .dm-badge }
[ :octicons-package-16: Releases ](https://github.com/codenio/dive-mcp/releases){ .dm-badge }
[ :octicons-beaker-16: dive v0.12–v0.13 ](dive-compatibility.md){ .dm-badge }
[ :octicons-plug-16: MCP compatible ](https://modelcontextprotocol.io){ .dm-badge }

</div>

</div>

<div class="grid cards" markdown>

-   :material-layers-triple:{ .lg .middle } __Layer analysis__

    ---

    Inspect image layers, digests, sizes, and build commands without leaving your editor.

    [:octicons-arrow-right-24: MCP tools](tools.md)

-   :material-chart-donut:{ .lg .middle } __Efficiency scoring__

    ---

    Get wasted-byte totals, efficiency scores, and CI-style pass/fail thresholds.

    [:octicons-arrow-right-24: CI check](tools.md#ci_check-thresholds)

-   :material-robot:{ .lg .middle } __Works in your assistant__

    ---

    Drop-in configs for Cursor, Claude Code, Codex, and Claude Desktop.

    [:octicons-arrow-right-24: Configuration](configuration.md)

</div>

## Quick start

!!! tip "Prerequisites"

    Install [dive](https://github.com/wagoodman/dive) and ensure Docker or Podman is
    running. `dive-mcp` shells out to your local `dive` binary — it is not bundled.

=== "Install"

    ```sh
    go install github.com/codenio/dive-mcp/cmd/dive-mcp@latest
    dive-mcp --version
    ```

=== "Cursor"

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

=== "Claude Code"

    ```sh
    claude mcp add dive-mcp -- dive-mcp
    ```

=== "Codex"

    Add to `~/.codex/config.toml`:

    ```toml
    [mcp_servers.dive-mcp]
    command = "dive-mcp"
    args = []
    ```

Restart your MCP client, then ask:

> Analyze `alpine:latest` with dive-mcp and tell me the efficiency score and top wasted files.

See [Configuration](configuration.md) for environment variables and
[example configs](https://github.com/codenio/dive-mcp/tree/main/examples).

## Tools

| Tool | Description |
|---|---|
| `analyze_image` | Image size, efficiency score, wasted bytes, layer count |
| `list_layers` | Per-layer digest, size, and build command |
| `get_wasted_space` | Top duplicated files across layers |
| `ci_check` | Pass/fail against efficiency and waste thresholds |

## Documentation

| Doc | Contents |
|---|---|
| [Configuration](configuration.md) | Client setup and environment variables |
| [MCP tools](tools.md) | MCP tool reference |
| [Development guide](development.md) | Build, test, pre-commit hooks |
| [Architecture](architecture.md) | How dive-mcp invokes and parses dive |
| [Dive compatibility](dive-compatibility.md) | Supported dive versions |
| [Releasing](releasing.md) | Maintainer release workflow |

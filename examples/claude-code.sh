#!/usr/bin/env bash
# Register dive-mcp with Claude Code.
#
# Requires: `dive-mcp` on PATH (see README for install instructions) and the
# `dive` CLI plus docker or podman installed.
set -euo pipefail

# Basic registration (stdio transport, PATH-resolved binary and dive path):
claude mcp add dive-mcp -- dive-mcp

# Example with explicit env vars (custom dive location / longer timeout):
# claude mcp add dive-mcp \
#   --env DIVE_MCP_DIVE_PATH=/usr/local/bin/dive \
#   --env DIVE_MCP_TIMEOUT=10m \
#   -- dive-mcp

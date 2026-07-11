# Contributing to dive-mcp

Thanks for your interest in contributing. This project is a thin MCP wrapper around the
[`dive`](https://github.com/wagoodman/dive) CLI — contributions that keep that boundary
clear and the server reliable are especially welcome.

## Ways to contribute

- Bug reports and reproduction steps
- Documentation and example client configs
- New or improved MCP tools (with tests)
- Fixes for `dive` JSON parsing or CI-threshold logic
- Compatibility updates when upstream `dive` releases change
- CI, release, and maintainer workflow improvements

If you are unsure whether an idea fits, open an issue first. Small, focused pull
requests are easier to review than large ones.

## Prerequisites

- Go 1.26+ (see `go.mod`)
- [`dive`](https://github.com/wagoodman/dive) on your `PATH`
- Docker or Podman (for integration / compatibility checks)
- [pre-commit](https://pre-commit.com/#install) for local git hooks (`pip install pre-commit` or `brew install pre-commit`)
- Optional: [`golangci-lint`](https://golangci-lint.run/) (pre-commit installs a pinned version automatically)

## Getting started

```sh
git clone https://github.com/codenio/dive-mcp.git
cd dive-mcp
make build
make test
dive-mcp --version
```

Fork the repo on GitHub, create a branch from `main`, and open a pull request when
ready. See [docs/development.md](docs/development.md) for build commands, pre-commit
hooks, testing, and project layout.

## Pull request checklist

Before opening a PR, please:

- [ ] Run `make hooks-install` once per clone (enables pre-commit)
- [ ] Run `make pre-commit` (or let the hook run on commit)
- [ ] Update README, docs, or examples if user-facing behavior changed
- [ ] Add or update tests for non-trivial logic changes
- [ ] Run dive compat checks if you touched `internal/dive/` or JSON parsing

Describe what changed, why, and how you tested it. Link any related issues.

## Code style

- Match existing naming, error wrapping (`fmt.Errorf("...: %w", err)`), and package
  layout.
- Prefer small, readable changes over abstractions.
- Comments should explain non-obvious behavior, not restate the code.
- MCP tool inputs/outputs should stay structured and documented via jsonschema tags
  where the SDK supports them.

## Commit messages

Use clear, imperative subject lines:

```
Add podman source validation to analyze_image
Fix ci_check default for highest_wasted_bytes
```

Reference issues in the body when applicable (`Fixes #12`).

## Releases

See [docs/releasing.md](docs/releasing.md). Contributors do not need to handle releases
unless asked.

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE) that covers this project.

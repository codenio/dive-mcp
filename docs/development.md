# Development

## Prerequisites

- Go 1.26+ (see `go.mod`)
- [`dive`](https://github.com/wagoodman/dive) on your `PATH`
- Docker or Podman (for integration / compatibility checks)
- [pre-commit](https://pre-commit.com/#install) (`pip install pre-commit` or `brew install pre-commit`)

## Commands

```sh
make build          # ./bin/dive-mcp
make install        # go install into $GOBIN / $GOPATH/bin
make run            # go run ./cmd/dive-mcp
make test           # unit tests
make lint           # golangci-lint, or go vet as fallback
make fmt            # gofmt + go fmt
make pre-commit     # run all pre-commit hooks
make hooks-install  # install pre-commit git hooks for this clone
make clean          # remove bin/, dist/, and site/
make docs           # serve docs locally (requires MkDocs)
make docs-build     # build static site to ./site
make release        # cross-compile darwin/linux amd64+arm64 into ./dist
```

## Pre-commit hooks

Hook definitions live in [`.pre-commit-config.yaml`](https://github.com/codenio/dive-mcp/blob/main/.pre-commit-config.yaml). Install
once per clone:

```sh
make hooks-install
```

Run checks manually anytime:

```sh
make pre-commit
```

## Project layout

```
cmd/dive-mcp/         main entrypoint, --version flag
internal/dive/         dive CLI runner + JSON types
internal/server/       MCP tool registration and handlers
examples/              ready-to-copy client configs
scripts/               dive install and compat verification
testdata/              fixtures and CI test image Dockerfile
```

Keep new code inside `internal/` unless it is a new command. The server should shell out
to `dive` rather than reimplementing dive's analysis logic.

## Testing

### Unit tests

```sh
make test
```

Parser and helper tests live in `internal/dive/`. Fixture data is in
`internal/dive/testdata/sample.json`. Regenerate when the expected JSON shape changes:

```sh
dive <image> --json internal/dive/testdata/sample.json
```

### Dive compatibility

CI verifies that multiple `dive` releases produce JSON `dive-mcp` can parse:

```sh
./scripts/install-dive.sh 0.13.1
./scripts/verify-dive-compat.sh ./.dive-bin/dive
```

See [dive-compatibility.md](dive-compatibility.md) for the full maintenance workflow.

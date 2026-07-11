# Releasing

Maintainers cut releases by pushing a semver tag (`v0.1.0`, `v0.2.0`, …).

The [Release workflow](../.github/workflows/release.yml) cross-compiles darwin/linux
amd64+arm64 binaries and attaches them to the GitHub release with SHA256 checksums.

```sh
git tag v0.1.0
git push origin v0.1.0
```

Users can install a specific version via:

```sh
go install github.com/codenio/dive-mcp/cmd/dive-mcp@v0.1.0
```

Contributors do not need to handle releases unless asked.

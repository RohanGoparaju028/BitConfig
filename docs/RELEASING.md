# Releasing BitConfig

This guide is for maintainers publishing a version of BitConfig on GitHub.

## 1. Prepare the release commit

- Confirm the README and CLI help match the behavior being released.
- Run `go test ./...` and `go build ./...`.
- Review `git status` and commit the intended changes to `main`.

## 2. Build release archives

From the repository root, run:

```bash
./build_release.sh v0.1.0
```

Pass the version you are releasing, including the `v` prefix. The script writes platform archives and `SHA256SUMS.txt` under `dist/`. It cross-compiles these targets:

- macOS Apple Silicon: `darwin-arm64`
- macOS Intel: `darwin-amd64`
- Linux x86_64: `linux-amd64`
- Linux ARM64: `linux-arm64`
- Windows x86_64: `windows-amd64`

Each archive contains the CLI binary, the README, license, and the `gnn/` Python sources and requirements. The GNN files are needed because the CLI locates `gnn/train.py` next to the installed executable.

Inspect the archive contents and checksums before uploading:

```bash
tar -tzf dist/bitconfig-v0.1.0-darwin-arm64.tar.gz
cat dist/SHA256SUMS.txt
```

## 3. Publish on GitHub

After CI passes on the release commit, push a new version tag:

```bash
git tag v0.1.1
git push origin v0.1.1
```

Use the actual new version; existing tags do not trigger a new build unless pushed as new refs. The Release workflow validates the version, runs Go checks, builds all five archives using the Go version in `go.mod`, verifies checksums, and creates a **draft** release with generated notes and all assets. No repository secret is needed beyond the automatic `GITHUB_TOKEN`; repository policy must allow Actions write access to contents.

Open the draft on the repository's Releases page, review the notes and assets, and publish it. For prereleases, select the prerelease checkbox before publishing. If a run fails after creating a draft, inspect the existing draft before rerunning: the workflow deliberately does not overwrite releases.

Local packaging requires Bash, Python 3, and `shasum`. For matching checksums, use identical source files and the same Go and Python/zlib versions as the build being reproduced. The script disables CGO and VCS metadata and normalizes archive order, timestamps, owners, and permissions. CI builds twice and compares archive checksums. The optional GNN dependency versions are not locked by this workflow.

## 4. After publishing

Open the release page and confirm all five archives and the checksum file are present. Follow the README install steps using a downloaded archive to verify the published instructions and asset names.

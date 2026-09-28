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

Push the release commit to GitHub, then create a release from that commit:

1. Open the repository's **Releases** page and choose **Draft a new release**.
2. Create a tag matching the archive version, such as `v0.1.0`, on the intended commit.
3. Add a short title and user-facing notes describing features, fixes, and known limitations.
4. Upload every `dist/*.tar.gz` archive and `dist/SHA256SUMS.txt`.
5. Publish the release.

Alternatively, with GitHub CLI authenticated and the tag already pushed:

```bash
gh release create v0.1.0 dist/bitconfig-v0.1.0-*.tar.gz dist/SHA256SUMS.txt \
  --title "BitConfig v0.1.0" \
  --notes-file RELEASE_NOTES.md
```

## 4. After publishing

Open the release page and confirm all five archives and the checksum file are present. Follow the README install steps using a downloaded archive to verify the published instructions and asset names.

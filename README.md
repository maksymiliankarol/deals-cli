# deals-cli

Command-line client for `deals-api`.

```bash
deals search red ferrari
deals search -s red ferrari      # run and save the query
deals search --saved 1           # run saved search #1
deals saved                      # list saved searches
deals buy --description "Looking to buy a red Ferrari" --email me@example.com --phone +15551234567
deals help
```

Saved searches are stored locally in `searches.json` inside your OS config dir
(`~/.config/deals-cli/` on Linux, `~/Library/Application Support/deals-cli/` on macOS,
`%AppData%\deals-cli\` on Windows). Nothing from `buy` is stored.

## Install

### Linux and macOS

```bash
curl -fsSL https://deals-api.metalogrom.com/install-cli | sh
```

This installs `deals` into `~/.local/bin`. If that directory isn't on your PATH, the script tells you. To install somewhere else:

```bash
curl -fsSL https://deals-api.metalogrom.com/install-cli | INSTALL_DIR=/usr/local/bin sh
```

(Writing to `/usr/local/bin` usually needs root, so use `sudo sh` at the end of that command.)

Check it worked:

```bash
deals version
```

To upgrade, run the same command again.

### Windows

Download `deals_windows_amd64.zip` from the
[latest release](https://github.com/maksymiliankarol/deals-cli/releases/latest),
unzip it, and put `deals.exe` in a folder on your PATH.

### With Go

```bash
go install github.com/maksymiliankarol/deals-cli/cmd/deals@latest
```

Make sure `$(go env GOPATH)/bin` is on your PATH.

### Configuration

Set `DEALS_API_URL` to use a different API (default: `https://deals-api.metalogrom.com`).

## Development

Run `deals-api` in another terminal (`pnpm dev` on http://localhost:8787), then:

```bash
export DEALS_API_URL=http://localhost:8787   # once per terminal session
go run ./cmd/deals search red ferrari
```

Without the export, the CLI talks to production.

### Checks (run before every commit)

```bash
go fmt ./...
go vet ./...
go build ./...
```

### Test the installed binary

```bash
go install ./cmd/deals     # puts `deals` in $(go env GOPATH)/bin
deals version
```

## Git workflow

Work directly on `main`.

```bash
git status                      # check what will be committed
go fmt ./... && go vet ./... && go build ./...
git add .
git commit -m "Add saved searches"
git push
```

Pushing does not publish a release. Only pushing a tag does (see Releasing below).

## Releasing

Versions follow [semver](https://semver.org): `vMAJOR.MINOR.PATCH`.

- patch (`v0.1.1`): bug fixes
- minor (`v0.2.0`): new commands or flags, backwards compatible
- major (`v1.0.0`): breaking changes

Steps:

```bash
# 1. be on an up-to-date, clean main
git switch main && git pull
git status                      # should be clean

# 2. run the checks
go fmt ./... && go vet ./... && go build ./...

# 3. (optional) build all release binaries locally into dist/, nothing is published
goreleaser release --snapshot --clean

# 4. tag and push; this triggers the GitHub Actions release workflow
git tag v0.0.1
git push origin v0.0.1
```

CI (`.github/workflows/release.yml`) then builds Linux, macOS and Windows
binaries with GoReleaser and attaches them, plus `checksums.txt`, to a new
GitHub Release. Watch it under the repo's **Actions** tab. When it's green,
the release appears under **Releases**.

Verify the published binary:

```bash
curl -L https://github.com/maksymiliankarol/deals-cli/releases/latest/download/deals_linux_amd64.tar.gz | tar xz
./deals version                 # should print the tag, e.g. deals 0.1.0
```

### If a release goes wrong

Never reuse a tag that has been published: the Go module proxy caches versions
permanently, so `go install` users would keep getting the old code.
Fix the problem and release the next patch version instead (`v0.1.1`).

If CI failed before anything was published, you can delete the tag and retry:

```bash
git tag -d v0.1.0
git push origin --delete v0.1.0
# fix, commit, then tag and push again
```

If a bad release is already public, also delete it on the GitHub Releases page
(or `gh release delete v0.1.0`) and ship the fixed version.

## Project layout

```
cmd/deals/
  main.go      command dispatch, help, version
  client.go    API URL, shared HTTP client, friendly API errors
  search.go    search command
  buy.go       buy command
  saved.go     saved searches (local JSON file)
.goreleaser.yaml
.github/workflows/release.yml
```
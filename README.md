# deals-cli

Command-line client for `deals-api`.

```bash
deals search red ferrari
deals search -s red ferrari      # run and save the query
deals search --saved 1           # run saved search #1
deals saved                      # list saved searches
deals unlock 1                   # get a payment link for listing #1, wait for payment
deals unlocked                   # list contacts you unlocked
deals buy --description "Looking to buy a red Ferrari" --email me@example.com --phone +15551234567
deals help
```

## Unlocking a contact

`deals unlock <id>` prints a Stripe payment link, then waits (up to 35 minutes, the
time the listing is held for you) until the payment goes through. The buyer's
contact details are then printed and saved locally.

The API shows the details **only once**, so the CLI saves them to disk before
printing them. If saving fails, they are still printed: write them down.

```bash
deals unlocked                   # list unlocked contacts
deals unlocked rm 2              # remove unlocked contact #2
```

If you press Ctrl+C while waiting, the link keeps working until it expires, but the
CLI can no longer show you the details. Run `deals unlock <id>` again only if you
haven't paid.

## Local files

Everything is stored in your OS config dir (`~/.config/deals-cli/` on Linux,
`~/Library/Application Support/deals-cli/` on macOS, `%AppData%\deals-cli\` on
Windows), readable only by your user:

- `searches.json`: saved searches
- `unlocked.json`: unlocked contacts (includes email and phone numbers)

Nothing from `buy` is stored.

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

To test `unlock` locally you also need Stripe forwarding to the API:

```bash
stripe listen --forward-to localhost:8787/stripe/webhook
```

Pay with the test card `4242 4242 4242 4242`, any future expiry, any CVC.

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
  saved.go     saved searches (local JSON file), shared config path helper
  unlock.go    unlock command: payment link, wait for payment, reveal
  unlocked.go  unlocked contacts (local JSON file)
.goreleaser.yaml
.github/workflows/release.yml
```
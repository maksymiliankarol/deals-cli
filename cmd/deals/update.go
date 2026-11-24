package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	repo          = "maksymiliankarol/deals-cli"
	checkInterval = 24 * time.Hour
)

// Downloads are a few MB, so they get more time than regular API calls.
var downloadClient = &http.Client{Timeout: 2 * time.Minute}

// latestVersion returns the newest release without the "v" prefix, e.g. "0.2.0".
func latestVersion(c *http.Client) (string, error) {
	resp, err := c.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checking latest version: %s", resp.Status)
	}
	var r struct {
		TagName string `json:"tag_name"`
	}
	if err := json.UnmarshalRead(resp.Body, &r); err != nil {
		return "", err
	}
	if r.TagName == "" {
		return "", errors.New("latest release has no tag")
	}
	return strings.TrimPrefix(r.TagName, "v"), nil
}

// newer reports whether version a is higher than b. Only MAJOR.MINOR.PATCH is
// compared; a "-rc1" style suffix is ignored.
func newer(a, b string) bool {
	pa, pb := versionParts(a), versionParts(b)
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] > pb[i]
		}
	}
	return false
}

func versionParts(v string) (p [3]int) {
	v, _, _ = strings.Cut(v, "-")
	for i, s := range strings.SplitN(v, ".", 3) {
		p[i], _ = strconv.Atoi(s)
	}
	return p
}

type updateCache struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
}

// checkForUpdate runs updateNotice in the background. Read the message from the
// returned channel once the command is done.
func checkForUpdate() <-chan string {
	ch := make(chan string, 1)
	go func() { ch <- updateNotice() }()
	return ch
}

// updateNotice returns a "new version" message, or "". It asks GitHub at most once
// per checkInterval and never fails the command. Set DEALS_NO_UPDATE_CHECK to disable.
func updateNotice() string {
	cur := currentVersion()
	if cur == "" || !isTerminal(os.Stderr) || os.Getenv("DEALS_NO_UPDATE_CHECK") != "" {
		return ""
	}
	path, err := configPath("update-check.json")
	if err != nil {
		return ""
	}

	var c updateCache
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &c)
	}
	if time.Since(c.CheckedAt) > checkInterval {
		if v, err := latestVersion(&http.Client{Timeout: 2 * time.Second}); err == nil {
			c.Latest = v
		}
		c.CheckedAt = time.Now() // also after a failure, so an offline user isn't delayed on every run
		if data, err := json.Marshal(c); err == nil {
			_ = os.MkdirAll(filepath.Dir(path), 0o700)
			_ = os.WriteFile(path, data, 0o600)
		}
	}

	if !newer(c.Latest, cur) {
		return ""
	}
	return fmt.Sprintf("\nA new version of deals is available: %s (you have %s). Run: deals update", c.Latest, cur)
}

// runUpdate downloads the latest release archive, checks it against the release's
// checksums.txt and replaces this binary with the one inside.
func runUpdate(args []string) error {
	if len(args) != 0 {
		return errors.New("usage: deals update")
	}
	if runtime.GOOS == "windows" {
		return fmt.Errorf("update isn't supported on Windows, download the zip from https://github.com/%s/releases/latest", repo)
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return err
	}

	done := status("Checking for updates...")
	latest, err := latestVersion(httpClient)
	done()
	if err != nil {
		return err
	}
	if !newer(latest, currentVersion()) {
		fmt.Printf("Already up to date (%s).\n", currentVersion())
		return nil
	}

	name := fmt.Sprintf("deals_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	base := fmt.Sprintf("https://github.com/%s/releases/download/v%s/", repo, latest)

	done = status("Downloading " + latest + "...")
	archive, err := download(base + name)
	var sums []byte
	if err == nil {
		sums, err = download(base + "checksums.txt")
	}
	done()
	if err != nil {
		return err
	}
	if err := verify(archive, sums, name); err != nil {
		return err
	}

	if err := installBinary(archive, exe); err != nil {
		return fmt.Errorf("update failed (no write access to %s? try sudo): %w", filepath.Dir(exe), err)
	}
	fmt.Printf("Updated to %s.\n", latest)
	return nil
}

func download(url string) ([]byte, error) {
	resp, err := downloadClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("downloading %s: %s", url, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 100<<20))
}

// verify checks the archive's sha256 against its line in checksums.txt
// ("<hash>  <file name>" per line, as written by GoReleaser).
func verify(archive, sums []byte, name string) error {
	sum := sha256.Sum256(archive)
	got := hex.EncodeToString(sum[:])
	for _, line := range strings.Split(string(sums), "\n") {
		if f := strings.Fields(line); len(f) == 2 && f[1] == name {
			if f[0] != got {
				return errors.New("checksum mismatch, update aborted")
			}
			return nil
		}
	}
	return fmt.Errorf("no checksum for %s, update aborted", name)
}

// installBinary extracts "deals" from the tar.gz archive and puts it at exe.
func installBinary(archive []byte, exe string) error {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return errors.New("no deals binary in the archive")
		}
		if err != nil {
			return err
		}
		if h.Name == "deals" && h.Typeflag == tar.TypeReg {
			return replaceFile(exe, tr)
		}
	}
}

// replaceFile writes r to a temp file next to dst and renames it over dst.
// The rename is atomic, so dst is never left half-written.
func replaceFile(dst string, r io.Reader) error {
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".deals-update-")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	_, err = io.Copy(tmp, r)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp.Name(), 0o755)
	}
	if err == nil {
		err = os.Rename(tmp.Name(), dst)
	}
	return err
}

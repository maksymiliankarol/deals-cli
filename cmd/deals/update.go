package main

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repo          = "maksymiliankarol/deals-cli"
	checkInterval = 24 * time.Hour
)

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

type updateCache struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
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

	if c.Latest == "" || c.Latest == cur {
		return ""
	}
	return fmt.Sprintf("\nA new version of deals is available: %s (you have %s). Run: deals update", c.Latest, cur)
}

// runUpdate re-runs install.sh, installing into the folder this binary lives in.
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
	if latest == currentVersion() {
		fmt.Printf("Already up to date (%s).\n", latest)
		return nil
	}

	resp, err := httpClient.Get("https://raw.githubusercontent.com/" + repo + "/main/install.sh")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading installer: %s", resp.Status)
	}

	// The script goes to sh's stdin and the folder travels as an env var,
	// so nothing is ever spliced into a shell command line.
	cmd := exec.Command("sh")
	cmd.Stdin = resp.Body
	cmd.Env = append(os.Environ(), "INSTALL_DIR="+filepath.Dir(exe))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("update failed (no write access to %s? try sudo): %w", filepath.Dir(exe), err)
	}
	return nil
}
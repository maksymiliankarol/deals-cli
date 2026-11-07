package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

func runUnlocked(args []string) error {
	switch {
	case len(args) == 0:
		return listUnlocked()
	case len(args) == 2 && args[0] == "rm":
		return removeUnlocked(args[1])
	default:
		return errors.New("usage: deals unlocked [rm <id>]")
	}
}

func listUnlocked() error {
	list, err := loadUnlocked()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Println("Nothing unlocked yet.")
		return nil
	}
	for _, u := range list {
		fmt.Printf("#%d  %s  %.2f %s\n    %s\n    email: %s\n    phone: %s\n\n",
			u.ID, u.UnlockedAt, float64(u.Price)/100, u.Currency, u.Description, u.Email, u.Phone)
	}
	return nil
}

// saveUnlocked stores the item newest first, replacing any older copy of the same id.
func saveUnlocked(item unlocked) error {
	list, err := loadUnlocked()
	if err != nil {
		return err
	}
	return writeUnlocked(append([]unlocked{item}, withoutID(list, item.ID)...))
}

func removeUnlocked(arg string) error {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return fmt.Errorf("invalid id %q", arg)
	}
	list, err := loadUnlocked()
	if err != nil {
		return err
	}
	next := withoutID(list, id)
	if len(next) == len(list) {
		return fmt.Errorf("no unlocked listing #%d (see: deals unlocked)", id)
	}
	return writeUnlocked(next)
}

func withoutID(list []unlocked, id int) []unlocked {
	var out []unlocked
	for _, u := range list {
		if u.ID != id {
			out = append(out, u)
		}
	}
	return out
}

func loadUnlocked() ([]unlocked, error) {
	path, err := configPath("unlocked.json")
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil // first run: no file yet
	}
	if err != nil {
		return nil, err
	}
	var list []unlocked
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("%s is corrupted: %w", path, err)
	}
	return list, nil
}

func writeUnlocked(list []unlocked) error {
	path, err := configPath("unlocked.json")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
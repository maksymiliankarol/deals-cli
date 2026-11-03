package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func runSaved(args []string) error {
	if len(args) == 0 {
		return listSaved()
	}
	switch args[0] {
	case "add":
		return addSaved(strings.Join(args[1:], " "))
	case "rm":
		if len(args) != 2 {
			return errors.New("usage: deals saved rm <n>")
		}
		return removeSaved(args[1])
	default:
		return errors.New("usage: deals saved [add <query> | rm <n>]")
	}
}

func listSaved() error {
	list, err := loadSaved()
	if err != nil {
		return err
	}
	if len(list) == 0 {
		fmt.Println("No saved searches.")
		return nil
	}
	for i, q := range list {
		fmt.Printf("%d. %s\n", i+1, q)
	}
	return nil
}

func addSaved(q string) error {
	q = strings.TrimSpace(q)
	// Same limits as the API's search schema, so we never save a query it would reject.
	if n := len([]rune(q)); n < 3 || n > 500 {
		return errors.New("query must be 3-500 characters")
	}

	list, err := loadSaved()
	if err != nil {
		return err
	}
	for _, s := range list {
		if strings.EqualFold(s, q) {
			fmt.Println("Already saved.")
			return nil
		}
	}
	if err := writeSaved(append(list, q)); err != nil {
		return err
	}
	fmt.Printf("Saved #%d\n", len(list)+1)
	return nil
}

func removeSaved(arg string) error {
	list, err := loadSaved()
	if err != nil {
		return err
	}
	n, err := strconv.Atoi(arg)
	if err != nil || n < 1 || n > len(list) {
		return fmt.Errorf("no saved search #%s (see: deals saved)", arg)
	}
	return writeSaved(append(list[:n-1], list[n:]...))
}

// savedQuery returns saved search #n (1-based).
func savedQuery(n int) (string, error) {
	list, err := loadSaved()
	if err != nil {
		return "", err
	}
	if n < 1 || n > len(list) {
		return "", fmt.Errorf("no saved search #%d (see: deals saved)", n)
	}
	return list[n-1], nil
}

func savedPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "deals-cli", "searches.json"), nil
}

func loadSaved() ([]string, error) {
	path, err := savedPath()
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
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("%s is corrupted: %w", path, err)
	}
	return list, nil
}

func writeSaved(list []string) error {
	path, err := savedPath()
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
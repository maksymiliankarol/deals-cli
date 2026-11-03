package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type result struct {
	ID          int     `json:"id"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
	Score       float64 `json:"score"`
}

func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ExitOnError)
	save := fs.Bool("s", false, "also save this search")
	n := fs.Int("saved", 0, "run saved search #n")
	fs.Parse(args)

	var q string
	if *n > 0 {
		if *save || fs.NArg() > 0 {
			return errors.New("--saved can't be combined with a query or -s")
		}
		var err error
		if q, err = savedQuery(*n); err != nil {
			return err
		}
	} else {
		q = strings.Join(fs.Args(), " ")
	}
	if strings.TrimSpace(q) == "" {
		return errors.New("usage: deals search [-s] <query> | deals search --saved <n>")
	}

	if err := search(q); err != nil {
		return err
	}
	if *save {
		return addSaved(q) // only after the API accepted the query
	}
	return nil
}

func search(q string) error {
	resp, err := httpClient.Get(apiURL() + "/search?q=" + url.QueryEscape(q))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return apiError(resp)
	}

	var results []result
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return err
	}
	if len(results) == 0 {
		fmt.Println("No matches.")
		return nil
	}
	for _, r := range results {
		fmt.Printf("#%d  (%.2f)  %s\n    %s\n\n", r.ID, r.Score, r.CreatedAt, r.Description)
	}
	return nil
}
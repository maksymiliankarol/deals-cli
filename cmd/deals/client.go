package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

func apiURL() string {
	if u := os.Getenv("DEALS_API_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://deals-api.metalogrom.com"
}

// apiMessage returns the "error" field of an API error body.
func apiMessage(body []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	_ = json.Unmarshal(body, &e)
	return e.Error
}

// apiError turns a non-success response into a readable error.
func apiError(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))

	if m := validationMessage(body); m != "" {
		return errors.New(m)
	}
	if m := apiMessage(body); m != "" {
		return errors.New(m)
	}
	return fmt.Errorf("%s: %s", resp.Status, body)
}

// validationMessage formats the validator's 400 body as one line per problem.
// It returns "" if the body isn't in the expected shape.
func validationMessage(body []byte) string {
	var v struct {
		Issues []struct {
			Message string `json:"message"`
			Path    []struct {
				Key string `json:"key"`
			} `json:"path"`
		} `json:"issues"`
	}
	if json.Unmarshal(body, &v) != nil || len(v.Issues) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("invalid input:")
	for _, i := range v.Issues {
		field := "input"
		if len(i.Path) > 0 {
			field = i.Path[0].Key
		}
		fmt.Fprintf(&b, "\n  - %s: %s", field, i.Message)
	}
	return b.String()
}
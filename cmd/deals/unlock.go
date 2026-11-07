package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	pollEvery   = 5 * time.Second
	pollTimeout = 35 * time.Minute // same as the API's reservation hold
)

type unlocked struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Price       int    `json:"price"`
	Currency    string `json:"currency"`
	ListedAt    string `json:"listed_at"`
	UnlockedAt  string `json:"unlocked_at"`
}

func runUnlock(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: deals unlock <id>")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil || id < 1 {
		return fmt.Errorf("invalid id %q", args[0])
	}

	sessionID, payURL, err := startUnlock(id)
	if err != nil {
		return err
	}

	fmt.Println("Open this link to pay:")
	fmt.Println(payURL)
	if openBrowser(payURL) == nil {
		fmt.Println("\n(Opened in your browser.)")
	}
	fmt.Println("\nWaiting for payment... (Ctrl+C to stop)")

	u, err := waitForContact(sessionID)
	if err != nil {
		return err
	}

	// Save first: the reveal is one-time, so never lose the details.
	saveErr := saveUnlocked(*u)

	fmt.Printf("\nUnlocked #%d\n  %s\n  email: %s\n  phone: %s\n", u.ID, u.Description, u.Email, u.Phone)

	if saveErr != nil {
		return fmt.Errorf("could not save these details, write them down now: %w", saveErr)
	}
	fmt.Println("\nSaved. See all with: deals unlocked")
	return nil
}

func startUnlock(id int) (sessionID, payURL string, err error) {
	resp, err := httpClient.Post(fmt.Sprintf("%s/unlock/%d?client=cli", apiURL(), id), "application/json", nil)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", "", apiError(resp)
	}

	var s struct {
		URL       string `json:"url"`
		SessionID string `json:"sessionId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return "", "", err
	}
	return s.SessionID, s.URL, nil
}

func waitForContact(sessionID string) (*unlocked, error) {
	deadline := time.Now().Add(pollTimeout)
	for time.Now().Before(deadline) {
		u, err := checkUnlock(sessionID)
		if err != nil || u != nil {
			return u, err
		}
		time.Sleep(pollEvery)
	}
	return nil, errors.New("timed out waiting for payment")
}

// checkUnlock asks the API once. It returns (nil, nil) while we should keep waiting.
func checkUnlock(sessionID string) (*unlocked, error) {
	resp, err := httpClient.Get(apiURL() + "/unlock/" + url.PathEscape(sessionID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var u unlocked
		if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
			return nil, err
		}
		return &u, nil
	case http.StatusAccepted, http.StatusTooManyRequests:
		return nil, nil
	case http.StatusGone:
		return nil, errors.New("these details were already revealed and can't be shown again")
	case http.StatusNotFound:
		return nil, errors.New("no payment found and the hold expired, run 'deals unlock <id>' again")
	}
	return nil, apiError(resp)
}
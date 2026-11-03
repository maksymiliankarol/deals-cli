package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
)

type buyRequest struct {
	Description string `json:"description"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
}

func runBuy(args []string) error {
	fs := flag.NewFlagSet("buy", flag.ExitOnError)
	var req buyRequest
	fs.StringVar(&req.Description, "description", "", "what you want to buy (10-500 chars)")
	fs.StringVar(&req.Email, "email", "", "your email")
	fs.StringVar(&req.Phone, "phone", "", "your phone, e.g. +15551234567")
	fs.Parse(args)

	if req.Description == "" || req.Email == "" || req.Phone == "" {
		return errors.New("usage: deals buy --description <text> --email <email> --phone <phone>")
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := httpClient.Post(apiURL()+"/buy", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return apiError(resp)
	}

	var created struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return err
	}
	fmt.Printf("Created listing #%d\n", created.ID)
	return nil
}
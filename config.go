package main

import (
	"encoding/json"
	"os"
)

type config struct {
	APIKeyGoogle string `json:"apiKeyGoogle"`
	APIKeyLoggly string `json:"apiKeyLoggly"`
	EntryDigit   int    `json:"entryDigit"`
	Host         string `json:"host"`
	MaxErrors    int    `json:"maxErrors"`
	MaxRetries   int    `json:"maxRetries"`
	Name         string `json:"name"`
	Redis        struct {
		DB       int    `json:"db"`
		Host     string `json:"host"`
		Password string `json:"password"`
		Port     int    `json:"port"`
	} `json:"redis"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Phone    string `json:"phone"`
	Timeout  int    `json:"timeout"`
}

func newConfig() *config {
	return &config{}
}

func (c *config) loadFromJSON(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(c); err != nil {
		return err
	}

	return nil
}

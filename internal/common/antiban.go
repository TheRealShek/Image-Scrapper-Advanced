package common

import (
	"math/rand"
	"net/http"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
}

// NewClient returns an http.Client with a random User-Agent and optional proxy.
func NewClient() *http.Client {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	return client
}

// RandomUserAgent returns a random User-Agent string.
func RandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

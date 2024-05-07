package main

import "errors"

type emoji string

var (
	Red   emoji = "🔴"
	Green emoji = "🟢"
)

var (
	ErrNotSuccess = errors.New("http response not 200")
)

// Response for the 'cluster/servers' API
type ServerLocation struct {
	Servers []ServerAddress `json:"servers,omitempty"`
	Leader  string          `json:"leader,omitempty"`
}

type ServerAddress struct {
	ID      string `json:"ID,omitempty"`
	Address string `json:"Address,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

package sdk

import (
	"io"
	"net/http"
	"time"
)

// APIResponse struct
type APIResponse struct {
	Headers    http.Header
	Body       []byte
	Status     string
	StatusCode int
	Location   string
}

// APIRequest struct
type APIRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Body    io.Reader
}

// APIError represents an error from the vRA API.
type APIError struct {
	Errors []Error `json:"errors"`
}

// Error API Error
type Error struct {
	Code          int    `json:"code"`
	Message       string `json:"message"`
	SystemMessage string `json:"systemMessage"`
}

// AuthResponse - This struct contains response of user authentication call.
type AuthResponse struct {
	Expires time.Time `json:"expires"`
	ID      string    `json:"id"`
	Tenant  string    `json:"tenant"`
}

// AuthenticationRequest represents the auth request to vra
type AuthenticationRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Tenant   string `json:"tenant"`
}

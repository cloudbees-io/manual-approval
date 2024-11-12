package manual_approval

import (
	"context"
	"net/http"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Config struct {
	context.Context
	Client HttpClient

	// Handler field allows you to handler.
	Handler string `json:"handler,omitempty"`
}

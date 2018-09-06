// Package egobee encapsulates types and helper functions for interacting with
// the ecobee REST API in Go.
package egobee

import (
	"fmt"
	"net/http"
)

// authorizingTransport is a RoundTripper which includes the Access token in the
// request headers as appropriate for accessing the ecobee API.
type authorizingTransport struct {
	auth      TokenStore
	transport http.RoundTripper
}

func (t *authorizingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", t.auth.AccessToken()))
	return t.transport.RoundTrip(req)
}

// Client for the ecobee API.
type Client struct {
	http.Client
}

// New egobee client.
func New(ts TokenStore) *Client {
	return &Client{
		Client: http.Client{
			Transport: &authorizingTransport{
				auth:      ts,
				transport: http.DefaultTransport,
			},
		},
	}
}

package egobee

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	authURLTemplate = `https://api.ecobee.com/authorize?response_type=ecobeePin&scope=smartWrite&client_id=%v`
	tokenURL        = "https://api.ecobee.com/token"
)

var errInvalidPinAuthenticator = errors.New("invalid PinAuthenticator")

// PinAuthenticationChallenge is the initial response from the Ecobee API for
// pin-based application authentication.
type PinAuthenticationChallenge struct {
	Pin               string `json:"ecobeePin"`
	AuthorizationCode string `json:"code"`
	Scope             Scope  `json:"scope"`
	// expires_in and interval are ignored for now.
}

// PinAuthenticator is a helper for handling the interative PIN authentication
// workflow.
type PinAuthenticator struct {
	appID string
	PinAuthenticationChallenge
}

// GetPin begines the PIN authentication workflow, and returns the PIN necessary
// for the user to interactively authenticate.
func (p *PinAuthenticator) GetPin() (string, error) {
	if p == nil || p.appID == "" {
		return "", errInvalidPinAuthenticator
	}
	resp, err := http.Get(fmt.Sprintf(authURLTemplate, p.appID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&p.PinAuthenticationChallenge); err != nil {
		return "", err
	}
	return p.Pin, nil
}

// Finalize indicates that the user has completed the PIN authentication, and
// initializes the provided TokenStorer with the initial access and refresh
// tokens.
func (p *PinAuthenticator) Finalize(ts TokenStorer) error {
	if p == nil || p.appID == "" || p.AuthorizationCode == "" {
		return errInvalidPinAuthenticator
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "grant_type=ecobeePin&code=%v&client_id=%v", p.AuthorizationCode, p.appID)

	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", &buf)
	r, err := reauthResponseFromHTTPResponse(resp)
	resp.Body.Close()
	if err != nil {
		return err
	}
	if !r.ok() {
		return r.err()
	}
	ts.Update(r.Resp)
	return nil
}

// NewPinAuthenticator gets a new PIN authentication helper.
func NewPinAuthenticator(appID string) *PinAuthenticator {
	return &PinAuthenticator{
		appID: appID,
	}
}

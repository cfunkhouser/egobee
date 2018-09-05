package egobee

import (
	"encoding/json"
	"errors"
	"regexp"
	"time"
)

var (
	// ErrInvalidDuration is returned from UnmarshalJSON when the JSON does not
	// represent a Duration.
	ErrInvalidDuration = errors.New("invalid duration")

	hasUnitRx = regexp.MustCompile("[a-zA-Z]+")
)

// Scope of a token.
type Scope string

// Possible Scopes.
// See https://www.ecobee.com/home/developer/api/documentation/v1/auth/auth-intro.shtml
var (
	ScopeSmartRead  Scope = "smartRead"
	ScopeSmartWrite Scope = "smartWrite"
	ScopeEMSWrite   Scope = "ems"
)

// TokenDuration wraps time.Duration to add JSON (un)marshalling
type TokenDuration struct {
	time.Duration
}

// MarshalJSON returns JSON representation of Duration.
func (d TokenDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// UnmarshalJSON returns a Duration from JSON representation. Since the ecobee
// API returns durations in Seconds, values will be treated as such.
func (d *TokenDuration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Second * time.Duration(value)
	case string:
		if !hasUnitRx.Match([]byte(value)) {
			value = value + "s"
		}
		dv, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		d.Duration = dv
	default:
		return ErrInvalidDuration
	}
	return nil
}

// TokenRefreshResponse is returned by the ecobee API on toke refresh.
// See https://www.ecobee.com/home/developer/api/documentation/v1/auth/token-refresh.shtml
type TokenRefreshResponse struct {
	AccessToken  string        `json:"access_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    TokenDuration `json:"expires_in"`
	RefreshToken string        `json:"refresh_token"`
	Scope        Scope         `json:"scope"`
}

// APIError returned by ecobee.
type APIError string

// Possible API Errors
var (
	APIErrorAccessDenied         APIError = "access_denied"
	APIErrorInvalidRequest       APIError = "invalid_request"
	APIErrorInvalidClient        APIError = "invalid_client"
	APIErrorInvalidGrant         APIError = "invalid_grant"
	APIErrorUnauthorizeClient    APIError = "unauthorized_client"
	APIErrorUnsupportedGrantType APIError = "unsupported_grant_type"
	APIErrorInvalidScope         APIError = "invalid_scope"
	APIErrorNotSupported         APIError = "not_supported"
	APIErrorAccountLocked        APIError = "account_locked"
	APIErrorAccountDisabled      APIError = "account_disabled"
	APIErrorAuthorizationPending APIError = "authorization_pending"
	APIErrorAuthorizationExpired APIError = "authorization_expired"
	APIErrorSlowDown             APIError = "slow_down"
)

// ErrorResponse as returned from the ecobee API.
type ErrorResponse struct {
	Error       APIError `json:"error"`
	Description string   `json:"error_description"`
	URI         string   `json:"error_uri"`
}

package egobee

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"sync"
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

// AuthorizationError returned by ecobee.
type AuthorizationError string

// Possible API Errors
var (
	AuthorizationErrorAccessDenied         AuthorizationError = "access_denied"
	AuthorizationErrorInvalidRequest       AuthorizationError = "invalid_request"
	AuthorizationErrorInvalidClient        AuthorizationError = "invalid_client"
	AuthorizationErrorInvalidGrant         AuthorizationError = "invalid_grant"
	AuthorizationErrorUnauthorizeClient    AuthorizationError = "unauthorized_client"
	AuthorizationErrorUnsupportedGrantType AuthorizationError = "unsupported_grant_type"
	AuthorizationErrorInvalidScope         AuthorizationError = "invalid_scope"
	AuthorizationErrorNotSupported         AuthorizationError = "not_supported"
	AuthorizationErrorAccountLocked        AuthorizationError = "account_locked"
	AuthorizationErrorAccountDisabled      AuthorizationError = "account_disabled"
	AuthorizationErrorAuthorizationPending AuthorizationError = "authorization_pending"
	AuthorizationErrorAuthorizationExpired AuthorizationError = "authorization_expired"
	AuthorizationErrorSlowDown             AuthorizationError = "slow_down"
)

// AuthorizationErrorResponse as returned from the ecobee API.
type AuthorizationErrorResponse struct {
	Error       AuthorizationError `json:"error"`
	Description string             `json:"error_description"`
	URI         string             `json:"error_uri"`
}

// Parse a response payload into the receiving AuthorizationErrorResponse. This will
// naturally fail if the payload is not an AuthorizationErrorResponse.
func (r *AuthorizationErrorResponse) Parse(payload []byte) error {
	if err := json.Unmarshal(payload, r); err != nil {
		return err
	}
	return nil
}

// ParseString behaves the same as Parse, but on a string.
func (r *AuthorizationErrorResponse) ParseString(payload string) error {
	return r.Parse([]byte(payload))
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

// TokenStorer for ecobee Access and Refresh tokens.
type TokenStorer interface {
	// AccessToken gets the access token from the store.
	GetAccessToken() (string, error)

	// RefreshToken gets the refresh token from the store.
	GetRefreshToken() (string, error)

	// ValidFor reports how much longer the access token is valid.
	GetValidFor() (time.Duration, error)

	// Update the TokenStorer with the contents of the response. This mutates the
	// access and refresh tokens.
	Update(*TokenRefreshResponse) error
}

// memoryStore implements tokenStore backed only by memory.
type memoryStore struct {
	mu           sync.RWMutex // protects the following members
	accessToken  string
	refreshToken string
	validUntil   time.Time
}

// persistentStore implements tokenStore backed by disk.
type persistentStore struct {
	mu           sync.RWMutex // protects the following members
	AccessToken  string
	RefreshToken string
	ValidUntil   time.Time
}

func (s *memoryStore) GetAccessToken() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessToken, nil
}

func (s *persistentStore) GetAccessToken() (string, error) {
	err := s.getPersistentTokenData()
	if err != nil {
		return "", err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AccessToken, err
}

func (s *memoryStore) GetRefreshToken() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refreshToken, nil
}

func (s *persistentStore) GetRefreshToken() (string, error) {
	err := s.getPersistentTokenData()
	if err != nil {
		return "", err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.RefreshToken, err
}

func (s *memoryStore) GetValidFor() (time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Now().Sub(s.validUntil), nil
}

func (s *persistentStore) GetValidFor() (time.Duration, error) {
	err := s.getPersistentTokenData()
	if err != nil {
		return 0, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Now().Sub(s.ValidUntil), err
}

func (s *memoryStore) Update(r *TokenRefreshResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = r.AccessToken
	s.refreshToken = r.RefreshToken
	s.validUntil = generateValidUntil(r)

	return nil
}

func (s *persistentStore) Update(r *TokenRefreshResponse) error {
	f, err := os.Create("/tmp/tokenStore")
	if err != nil {
		return err
	}
	defer f.Close()

	// Update in-memory data
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AccessToken = r.AccessToken
	s.RefreshToken = r.RefreshToken
	s.ValidUntil = generateValidUntil(r)

	// Write token data to file to be accessed later
	encoder := gob.NewEncoder(f)
	err = encoder.Encode(s)

	return err
}

// NewMemoryTokenStore is a TokenStorer with no persistence.
func NewMemoryTokenStore(r *TokenRefreshResponse) TokenStorer {
	s := &memoryStore{}
	s.Update(r)
	return s
}

// NewPersistentTokenStore is a ToeknStorer with persistence to disk
func NewPersistentTokenStore(r *TokenRefreshResponse) (TokenStorer, error) {
	s := &persistentStore{}
	// update persistent storage
	if err := s.Update(r); err != nil {
		return nil, err
	}
	return s, nil
}

// generateValidUntil returns the time the token expires with an added buffer
func generateValidUntil(r *TokenRefreshResponse) time.Time {
	// Subtract a few seconds to allow for network and processing delays.
	return time.Now().Add(r.ExpiresIn.Duration - (15 * time.Second))
}

// getPersistentTokenData returns the token data stored in a local file
func (s *persistentStore) getPersistentTokenData() error {
	f, err := os.Open("/tmp/tokenStore")
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(f)
	err = decoder.Decode(s)

	return err
}

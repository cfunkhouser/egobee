package egobee

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

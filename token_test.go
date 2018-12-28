package egobee

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
)

type testJSON struct {
	Duration TokenDuration `json:"duration"`
}

func TestUnmarshalTokenDuration(t *testing.T) {
	for _, tt := range []struct {
		name string
		json string
		want *testJSON
	}{
		{
			name: "unmarshal string duration with no units",
			json: `{"duration":"12345"}`,
			want: &testJSON{Duration: TokenDuration{Duration: time.Second * 12345}},
		},
		{
			name: "unmarshal float duration",
			json: `{"duration":12345}`,
			want: &testJSON{Duration: TokenDuration{Duration: time.Second * 12345}},
		},
		{
			name: "unmarshal string duration with units",
			json: `{"duration":"3h25m45s"}`,
			want: &testJSON{Duration: TokenDuration{Duration: time.Second * 12345}},
		},
	} {
		got := &testJSON{}
		if err := json.Unmarshal([]byte(tt.json), &got); err != nil {
			t.Errorf("%v: got unexpected error: %v", tt.name, err)
		} else if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%v: got: %v, wanted: %v", tt.name, got, tt.want)
		}
	}
}

func TestMarshalTokenDuration(t *testing.T) {
	for _, tt := range []struct {
		name string
		val  *testJSON
		want string
	}{
		{
			name: "marshal",
			val:  &testJSON{Duration: TokenDuration{Duration: time.Second * 12345}},
			want: `{"duration":"3h25m45s"}`,
		},
	} {
		if got, err := json.Marshal(tt.val); err != nil {
			t.Errorf("%v: got unexpected error: %v", tt.name, err)
		} else if string(got) != tt.want {
			t.Errorf("%v: got: %q, wanted: %q", tt.name, got, tt.want)
		}
	}
}

// TestUnmarshalTokenRefreshResponse tests that the example JSON provided on the
// ecobee API documentation page is properly unmarshalled.
// See https://www.ecobee.com/home/developer/api/documentation/v1/auth/token-refresh.shtml
func TestUnmarshalTokenRefreshResponse(t *testing.T) {
	// This JSON has stray whitespace which is preserved from the source docs.
	exampleJSON := `{
    "access_token": "Rc7JE8P7XUgSCPogLOx2VLMfITqQQrjg",
    "token_type": "Bearer",
    "expires_in": 3599,
    "refresh_token": "og2Obost3ucRo1ofo0EDoslGltmFMe2g",
    "scope": "smartWrite" 
}                `
	want := &TokenRefreshResponse{
		AccessToken:  "Rc7JE8P7XUgSCPogLOx2VLMfITqQQrjg",
		TokenType:    "Bearer",
		ExpiresIn:    TokenDuration{Duration: time.Second * 3599},
		RefreshToken: "og2Obost3ucRo1ofo0EDoslGltmFMe2g",
		Scope:        ScopeSmartWrite,
	}

	got := &TokenRefreshResponse{}
	if err := json.Unmarshal([]byte(exampleJSON), got); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, wanted: %+v", got, want)
	}
}

func TestUnmarshalAuthorizationErrorResponse(t *testing.T) {
	// This JSON has stray whitespace which is preserved from the source docs.
	exampleJSON := `{
    "error": "invalid_client",
    "error_description": "Authentication error, invalid authentication method, lack of credentials, etc.",
    "error_uri": "https://tools.ietf.org/html/rfc6749#section-5.2"
}`
	want := &AuthorizationErrorResponse{
		Error:       AuthorizationErrorInvalidClient,
		Description: "Authentication error, invalid authentication method, lack of credentials, etc.",
		URI:         "https://tools.ietf.org/html/rfc6749#section-5.2",
	}

	got := &AuthorizationErrorResponse{}
	if err := json.Unmarshal([]byte(exampleJSON), got); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, wanted: %+v", got, want)
	}
}

func TestAuthorizationErrorResponse_ParseString(t *testing.T) {
	exampleJSON := `{
    "error": "invalid_client",
    "error_description": "Authentication error, invalid authentication method, lack of credentials, etc.",
    "error_uri": "https://tools.ietf.org/html/rfc6749#section-5.2"
}`

	want := &AuthorizationErrorResponse{
		Error:       AuthorizationErrorInvalidClient,
		Description: "Authentication error, invalid authentication method, lack of credentials, etc.",
		URI:         "https://tools.ietf.org/html/rfc6749#section-5.2",
	}

	got := &AuthorizationErrorResponse{}
	if err := got.ParseString(exampleJSON); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, wanted: %+v", got, want)
	}
}

func TestAuthorizationErrorResponse_Parse(t *testing.T) {
	exampleJSON := []byte(`{
    "error": "invalid_client",
    "error_description": "Authentication error, invalid authentication method, lack of credentials, etc.",
    "error_uri": "https://tools.ietf.org/html/rfc6749#section-5.2"
}`)

	want := &AuthorizationErrorResponse{
		Error:       AuthorizationErrorInvalidClient,
		Description: "Authentication error, invalid authentication method, lack of credentials, etc.",
		URI:         "https://tools.ietf.org/html/rfc6749#section-5.2",
	}

	got := &AuthorizationErrorResponse{}
	if err := got.Parse(exampleJSON); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, wanted: %+v", got, want)
	}
}

func TestNewPersistentTokenStorer(t *testing.T) {
	testStorePath := "/tmp/tokenStore"
	testFileData := []byte(`{"accessToken":"anAccessToken","refreshToken":"aRefreshToken","validUntil":"2015-02-23T14:51:00.000000000-05:00"}`)
	err := ioutil.WriteFile(testStorePath, testFileData, 0640)
	tokenStore, err := NewPersistentTokenStorer(testStorePath)
	if err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if tokenStore.AccessToken() != "anAccessToken" {
		t.Errorf("access tokens do not match: %v vs. %v", tokenStore.AccessToken(), "anAccessToken")
	}
	if tokenStore.RefreshToken() != "aRefreshToken" {
		t.Errorf("refresh tokens do not match: %v vs. %v", tokenStore.RefreshToken(), "aRefreshToken")
	}
	if err := os.Remove(testStorePath); err != nil {
		t.Fatalf("Failed to remove temporary file: %v", err)
	}
}

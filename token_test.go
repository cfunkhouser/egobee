package egobee

import (
	"encoding/json"
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
		t.Errorf("unmarshal sample json: got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unmarshal sample json: got: %+v, wanted: %+v", got, want)
	}
}

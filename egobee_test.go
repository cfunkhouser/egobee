package egobee

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeTokenStorer struct {
	access  string
	refresh string
	vf      time.Duration
}

func (s *fakeTokenStorer) AccessToken() string {
	return s.access
}

func (s *fakeTokenStorer) RefreshToken() string {
	return s.refresh
}

func (s *fakeTokenStorer) ValidFor() time.Duration {
	return s.vf
}

func (s *fakeTokenStorer) Update(r *TokenRefreshResponse) error {
	return nil
}

func TestAPIBaseURL(t *testing.T) {
	abu := apiBaseURL("http://foo")
	want := "http://foo/bar/baz"
	if got := abu.URL("/bar/baz"); got != want {
		t.Errorf("got: %q, want: %q", got, want)
	}
}

func TestAuthorizingTransport(t *testing.T) {
	clientForTest := http.Client{
		Transport: &authorizingTransport{
			auth:      &fakeTokenStorer{"thisisanaccesstoken", "thisisarefreshtoken", time.Minute * 30},
			transport: http.DefaultTransport,
		},
	}
	serverForTest := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Authorization")
		if got != "Bearer thisisanaccesstoken" {
			t.Errorf(`invalide Authorization header; got: %q, want: "Bearer thisisanaccesstoken"`, got)
		}
	}))
	defer serverForTest.Close()
	res, err := clientForTest.Get(serverForTest.URL)
	if err != nil {
		t.Errorf("unexpected error GETting from test server: %v", err)
	}
	res.Body.Close()
}

func TestReauthResponse_OK(t *testing.T) {
	for _, tt := range []struct {
		name string
		resp *reauthResponse
		want bool
	}{
		{
			name: "response only",
			resp: &reauthResponse{
				Resp: &TokenRefreshResponse{},
			},
			want: true,
		},
		{
			// This should be impossible, but we'll test it anyway.
			name: "response and error",
			resp: &reauthResponse{
				Err:  &AuthorizationErrorResponse{},
				Resp: &TokenRefreshResponse{},
			},
			want: false,
		},
		{
			name: "error only",
			resp: &reauthResponse{
				Err: &AuthorizationErrorResponse{},
			},
			want: false,
		},
		{
			name: "empty non-nil receiver (zero)",
			resp: &reauthResponse{},
			want: false,
		},
		{
			name: "nil receiver",
			want: false,
		},
	} {
		if got := tt.resp.ok(); got != tt.want {
			t.Errorf("%v: got %v, wanted %v", tt.name, got, tt.want)
		}
	}
}

func TestAuthorizingTransport_ShouldReauth(t *testing.T) {
	for _, tt := range []struct {
		name string
		ts   TokenStorer
		want bool
	}{
		{
			name: "shouldn't reauth",
			ts:   &fakeTokenStorer{"foo", "bar", time.Minute * 30},
			want: false,
		},
		{
			name: "reauth for time",
			ts:   &fakeTokenStorer{"foo", "bar", time.Second},
			want: true,
		},
		{
			name: "reauth for token",
			ts:   &fakeTokenStorer{"", "", time.Minute * 30},
			want: true,
		},
		{
			name: "reauth for both", // just for good measure.
			ts:   &fakeTokenStorer{"", "", time.Second},
			want: true,
		},
	} {
		testTransport := &authorizingTransport{auth: tt.ts}
		if got := testTransport.shouldReauth(); got != tt.want {
			t.Errorf("%v: got %v, wanted %v", tt.name, got, tt.want)
		}
	}
}

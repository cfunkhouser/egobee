package egobee

import (
	"net/http"
	"net/http/httptest"
	"reflect"
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

func TestOptions_APIHost(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []apiBaseURL
		opt  *Options
		want apiBaseURL
	}{
		{
			name: "nil opts, no args",
			want: ecobeeAPIHost,
		},
		{
			name: "nil opts, one arg",
			args: []apiBaseURL{"https://api.foo.bar"},
			want: apiBaseURL("https://api.foo.bar"),
		},
		{
			name: "nil opts, multiple args",
			args: []apiBaseURL{"https://api.foo.bar", "https://api.example.com", "https://bar.foo.bar"},
			want: apiBaseURL("https://bar.foo.bar"),
		},
		{
			name: "empty APIHost, no args",
			opt:  &Options{},
			want: ecobeeAPIHost,
		},
		{
			name: "empty APIHost, one arg",
			opt:  &Options{},
			args: []apiBaseURL{"https://api.foo.bar"},
			want: apiBaseURL("https://api.foo.bar"),
		},
		{
			name: "empty APIHost, multiple args",
			opt:  &Options{},
			args: []apiBaseURL{"https://api.foo.bar", "https://api.example.com", "https://bar.foo.bar"},
			want: apiBaseURL("https://bar.foo.bar"),
		},
		{
			name: "set APIHost, no args",
			opt:  &Options{APIHost: "http://api.something.lol"},
			want: apiBaseURL("http://api.something.lol"),
		},
		{
			name: "set APIHost, one arg",
			opt:  &Options{APIHost: "http://api.something.lol"},
			args: []apiBaseURL{"https://api.foo.bar"},
			want: apiBaseURL("http://api.something.lol"),
		},
		{
			name: "set APIHost, multiple args",
			opt:  &Options{APIHost: "http://api.something.lol"},
			args: []apiBaseURL{"https://api.foo.bar", "https://api.example.com", "https://bar.foo.bar"},
			want: apiBaseURL("http://api.something.lol"),
		},
	} {
		if got := tt.opt.apiHost(tt.args...); got != tt.want {
			t.Errorf("%v: apiHost returned incorrect value; got: %v, want: %v", tt.name, got, tt.want)
		}
	}
}

func TestAccumulateOptions(t *testing.T) {
	for _, tt := range []struct {
		name string
		opts []*Options
		want *Options
	}{
		{
			name: "empty opts list",
			// Empty options list means nil *Options. Set neither here.
		},
		{
			name: "single opts entry",
			opts: []*Options{&Options{APIHost: "https://api.foo.bar"}},
			want: &Options{APIHost: "https://api.foo.bar"},
		},
		{
			name: "multiple opts entries with non-conflicting APIHost",
			opts: []*Options{
				&Options{APIHost: "https://api.foo.bar"},
				nil,
			},
			want: &Options{APIHost: "https://api.foo.bar"},
		},
		{
			name: "multiple opts entries with conflicting APIHost",
			opts: []*Options{
				&Options{APIHost: "https://api.foo.bar"},
				&Options{APIHost: "https://api.example.com"},
				&Options{APIHost: "https://bar.foo.bar"},
			},
			want: &Options{APIHost: "https://bar.foo.bar"},
		},
	} {
		got := accumulateOptions(tt.opts)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%v: accumulated options don't match;\n\tgot: %+v\n\t%+v", tt.name, got, tt.want)
		}
	}
}

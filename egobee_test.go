package egobee

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeTokenStore struct {
}

func (s *fakeTokenStore) AccessToken() string {
	return "thisisanaccesstoken"
}

func (s *fakeTokenStore) RefreshToken() string {
	return "thisisarefreshtoken"
}

func (s *fakeTokenStore) ValidFor() time.Duration {
	return time.Minute * 30
}

func TestAuthorizingTransport(t *testing.T) {
	clientForTest := http.Client{
		Transport: &authorizingTransport{
			auth:      &fakeTokenStore{},
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

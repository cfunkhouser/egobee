package egobee

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeTokenStorer struct {
}

func (s *fakeTokenStorer) GetAccessToken() (string, error) {
	return "thisisanaccesstoken", nil
}

func (s *fakeTokenStorer) GetRefreshToken() (string, error) {
	return "thisisarefreshtoken", nil
}

func (s *fakeTokenStorer) GetValidFor() (time.Duration, error) {
	return time.Minute * 30, nil
}

func (s *fakeTokenStorer) Update(r *TokenRefreshResponse) error {
	return nil
}

func TestAuthorizingTransport(t *testing.T) {
	clientForTest := http.Client{
		Transport: &authorizingTransport{
			auth:      &fakeTokenStorer{},
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

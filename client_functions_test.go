package egobee

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAssembleSelectURL(t *testing.T) {
	testAPIURL := "http://heylookathing"
	testSelection := &Selection{
		SelectionType:  SelectionTypeRegistered,
		SelectionMatch: "awwyiss",
	}

	want := "http://heylookathing?json=%7B%22selection%22%3A%7B%22selectionType%22%3A%22registered%22%2C%22selectionMatch%22%3A%22awwyiss%22%7D%7D"

	got, err := assembleSelectURL(testAPIURL, testSelection)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got != want {
		t.Fatalf("got: %+v, want: %v", got, want)
	}
}

func TestClientThermostatSummary(t *testing.T) {
	testValidResponse := `{
		"revisionList": ["revision1","revision2"],
		"thermostatCount": 2,
		"statusList": ["status1","status2"],
		"status": {"code": 200, "message": "Ok"}
	}`
	serverForTest := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != requestContentType {
			t.Errorf("invalid Content-Type header; got: %q, want: %q", got, requestContentType)
		}
		if got := r.URL.Path; got != thermostatSummaryURL {
			t.Errorf("invalid API Path; got: %q, want: %q", got, thermostatSummaryURL)
		}
		w.Header().Set("Content-Type", requestContentType)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testValidResponse))
	}))
	defer serverForTest.Close()

	clientForTest := &Client{
		api: apiBaseURL(serverForTest.URL),
	}

	want := &ThermostatSummary{
		RevisionList:    []string{"revision1", "revision2"},
		ThermostatCount: 2,
		StatusList:      []string{"status1", "status2"},
		Status: struct {
			Code    int    `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
		}{
			200,
			"Ok",
		},
	}
	got, err := clientForTest.ThermostatSummary()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("return value check failed;\ngot: %+v\nwant: %+v", got, want)
	}
}

package egobee

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestAssembleSelectionURL(t *testing.T) {
	testAPIURL := "http://heylookathing"
	testSelection := &Selection{
		SelectionType:  SelectionTypeRegistered,
		SelectionMatch: "awwyiss",
	}

	want := "http://heylookathing?json=%7B%22selection%22%3A%7B%22selectionType%22%3A%22registered%22%2C%22selectionMatch%22%3A%22awwyiss%22%7D%7D"
	got, err := assembleSelectionURL(testAPIURL, testSelection)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got != want {
		t.Fatalf("got: %+v, want: %v", got, want)
	}
}

func TestAssembleSelectionURLWhenMarshalingFails(t *testing.T) {
	origJSONMarshal := jsonMarshal
	jsonMarshal = func(_ interface{}) ([]byte, error) {
		return nil, errors.New("test error")
	}
	defer func() { jsonMarshal = origJSONMarshal }()

	got, err := assembleSelectionURL("", &Selection{})
	if got != "" {
		t.Errorf(`invalid return value; wanted: "", got: %v`, got)
	}
	if err == nil {
		t.Error("invalid error return value; wanted error, got nil")
	}
}

func TestAssembleSelectionRequest(t *testing.T) {
	testAPIURL := "http://heylooksomethingelse/endpoint"
	testSelection := &Selection{
		SelectionType:  SelectionTypeRegistered,
		SelectionMatch: "awwyiss",
	}

	got, err := assembleSelectionRequest(testAPIURL, testSelection)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if got.Method != http.MethodGet {
		t.Errorf("invalid method on request; got: %q, want: %q", got.Method, http.MethodGet)
	}
	wantu, _ := assembleSelectionURL(testAPIURL, testSelection)
	if gotu := got.URL.String(); gotu != wantu {
		t.Errorf("invalid URL; got: %q, want: %q", gotu, wantu)
	}
	if goth := got.Header.Get("Content-Type"); goth != requestContentType {
		t.Errorf("invalid Content-Type header; got: %q, want: %q", goth, requestContentType)
	}
}

func TestAssembleSelectionRequestWhenNewRequestFails(t *testing.T) {
	origHTTPNewRequest := httpNewRequest
	httpNewRequest = func(_, _ string, _ io.Reader) (*http.Request, error) {
		return nil, errors.New("test error")
	}
	defer func() { httpNewRequest = origHTTPNewRequest }()

	got, err := assembleSelectionRequest("", &Selection{})
	if got != nil {
		t.Errorf("got non-nil request: %+v", got)
	}
	if err == nil {
		t.Error("invalid error return value; wanted error, got nil")
	}
	if !strings.HasPrefix(err.Error(), "failed to create request:") {
		t.Errorf(`invalid error return value; wanted "failed to create request:" prefix, got: %q`, err)
	}
}

func TestValidateSelectionResponse(t *testing.T) {
	for _, tt := range []struct {
		res  *http.Response
		want error
	}{
		{
			res: &http.Response{
				Status:     "Found",
				StatusCode: http.StatusFound,
			},
			want: errors.New("non-ok status response from API: 302 Found"),
		},
		{
			res: &http.Response{
				Status:     "WTF Is This?",
				StatusCode: 600,
			},
			want: errors.New("non-ok status response from API: 600 WTF Is This?"),
		},
		{
			res: &http.Response{
				Status:     "OK",
				StatusCode: 200,
			},
		},
		{
			res: &http.Response{
				Status:     "Created",
				StatusCode: 201,
			},
		},
	} {
		if got := validateSelectionResponse(tt.res); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("invalid error response; got: %q, want: %q", got, tt.want)
		}
	}
}

func selectionRequestValidatingTestHandler(t *testing.T, testPayload string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != requestContentType {
			t.Errorf("invalid Content-Type header; got: %q, want: %q", got, requestContentType)
		}
		if got := r.URL.Path; got != thermostatSummaryURL {
			t.Errorf("invalid API Path; got: %q, want: %q", got, thermostatSummaryURL)
		}
		w.Header().Set("Content-Type", requestContentType)
		w.WriteHeader(http.StatusOK)
		if testPayload != "" {
			w.Write([]byte(testPayload))
		}
	}
}

func clientAndServerForTest(t *testing.T, testPayload string) (*Client, *httptest.Server) {
	s := httptest.NewServer(selectionRequestValidatingTestHandler(t, testPayload))
	c := &Client{
		api: apiBaseURL(s.URL),
	}
	return c, s
}

func TestClientThermostatSummary(t *testing.T) {
	testValidResponse := `{
		"revisionList": ["revision1","revision2"],
		"thermostatCount": 2,
		"statusList": ["status1","status2"],
		"status": {"code": 200, "message": "Ok"}
	}`

	client, server := clientAndServerForTest(t, testValidResponse)
	defer server.Close()

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
	got, err := client.ThermostatSummary()
	if err != nil {
		t.Fatalf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("return value check failed;\ngot: %+v\nwant: %+v", got, want)
	}
}

func TestClientThermostats(t *testing.T) {
	for _, tt := range []struct {
		name     string
		response string
		want     []*Thermostat
	}{
		{
			name: "OK response with thermostats",
			response: `{
		"page": {
			"page": 1,
			"totalPages": 1,
			"pageSize": 2,
			"total": 2
		},
		"thermostatList": [
			{ "name": "thermostat1" },
			{ "name": "thermostat2" }
		],
		"status": { "code": 200, "message": "OK" }
	}`,
			want: []*Thermostat{
				&Thermostat{Name: "thermostat1"},
				&Thermostat{Name: "thermostat2"},
			},
		},
		{
			name: "response with thermostats and no page info",
			response: `{
		"thermostatList": [
			{ "name": "thermostat1" },
			{ "name": "thermostat2" }
		],
		"status": { "code": 200, "message": "OK" }
	}`,
			want: []*Thermostat{
				&Thermostat{Name: "thermostat1"},
				&Thermostat{Name: "thermostat2"},
			},
		},
		{
			name: "OK response with empty thermostat list",
			response: `{
		"page": {
			"page": 1,
			"totalPages": 1,
			"pageSize": 2,
			"total": 2
		},
		"thermostatList": [],
		"status": { "code": 200, "message": "OK" }
	}`,
			want: []*Thermostat{},
		},
		{
			name: "OK response with no thermostat list",
			response: `{
		"page": {
			"page": 1,
			"totalPages": 1,
			"pageSize": 2,
			"total": 2
		},
		"status": { "code": 200, "message": "OK" }
	}`,
		},
	} {
		client, server := clientAndServerForTest(t, tt.response)
		got, err := client.Thermostats(&Selection{})
		if err != nil {
			t.Fatalf("case %q: got unexpected error: %v", tt.name, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("case %q: return value check failed;\ngot: %#v\nwant: %#v", tt.name, got, tt.want)
		}
		server.Close()
	}
}

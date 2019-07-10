package egobee

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/imdario/mergo"
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

type testServerOpts struct {
	Payload    string
	StatusCode int

	APIPath string
}

var defaultOpts testServerOpts

func clientAndServerForTest(t *testing.T, opts testServerOpts) (*Client, *httptest.Server) {
	t.Helper()
	s := httptest.NewServer(selectionRequestValidatingTestHandler(t, opts))
	c := &Client{
		api: apiBaseURL(s.URL),
	}
	return c, s
}

func selectionRequestValidatingTestHandler(t *testing.T, opts testServerOpts) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != requestContentType {
			t.Errorf("invalid Content-Type header; got: %q, want: %q", got, requestContentType)
		}
		if got := r.URL.Path; got != opts.APIPath {
			t.Errorf("invalid API Path; got: %q, want: %q", got, opts.APIPath)
		}
		w.Header().Set("Content-Type", requestContentType)
		statusCode := http.StatusOK
		if opts.StatusCode != 0 {
			statusCode = opts.StatusCode
		}
		w.WriteHeader(statusCode)
		if opts.Payload != "" {
			w.Write([]byte(opts.Payload))
		}
	}
}

func TestClientThermostatSummary(t *testing.T) {
	baseOpts := testServerOpts{
		APIPath: "/1/thermostatSummary",
	}
	for _, tt := range []struct {
		name    string
		opts    testServerOpts
		want    *ThermostatSummary
		wantErr string
	}{
		{
			name: "OK response",
			opts: testServerOpts{
				Payload: `{
		"revisionList": ["revision1","revision2"],
		"thermostatCount": 2,
		"statusList": ["status1","status2"],
		"status": {"code": 200, "message": "Ok"}
	}`,
			},
			want: &ThermostatSummary{
				RevisionList:    []string{"revision1", "revision2"},
				ThermostatCount: 2,
				StatusList:      []string{"status1", "status2"},
				Status: struct {
					Code    int    `json:"code,omitempty"`
					Message string `json:"message,omitempty"`
				}{200, "Ok"},
			},
		},
		{
			name: "Not-ok (503) response",
			opts: testServerOpts{
				StatusCode: 503,
			},
			wantErr: "non-ok status response from API: 503 Internal Server Error",
		},
	} {
		opts := baseOpts
		if err := mergo.Merge(&opts, tt.opts, mergo.WithOverride); err != nil {
			t.Fatalf("failed setting up opts for test: %v", err)
		}
		client, server := clientAndServerForTest(t, opts)
		got, err := client.ThermostatSummary()
		if tt.wantErr != "" && err == nil {
			t.Errorf("case %q: didn't get expected error; want: %q", tt.name, tt.wantErr)
		}
		if tt.wantErr == "" && err != nil {
			t.Errorf("case %q: got unexpected error: %v", tt.name, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("case %q: return value check failed;\ngot: %#v\nwant: %#v", tt.name, got, tt.want)
		}
		server.Close()
	}
}

func TestClientThermostatSummaryJSONDecodeError(t *testing.T) {
	origJSONDecode := jsonDecode
	jsonDecode = func(io.Reader, interface{}) error {
		return errors.New("test error")
	}
	defer func() { jsonDecode = origJSONDecode }()

	client, server := clientAndServerForTest(t, testServerOpts{APIPath: "/1/thermostatSummary"})
	defer server.Close()
	got, err := client.ThermostatSummary()

	if got != nil {
		t.Errorf("got unexpected return value; got: %+v, want: nil", got)
	}
	if err.Error() != "test error" {
		t.Errorf(`got unexpected error value; got: %v, want: "test error"`, err)
	}
}

func TestClientThermostats(t *testing.T) {

	baseOpts := testServerOpts{
		APIPath: "/1/thermostat",
	}

	for _, tt := range []struct {
		name    string
		opts    testServerOpts
		want    []*Thermostat
		wantErr string
	}{
		{
			name: "OK response with thermostats",
			opts: testServerOpts{
				Payload: `{
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
			},
			want: []*Thermostat{
				&Thermostat{Name: "thermostat1"},
				&Thermostat{Name: "thermostat2"},
			},
		},
		{
			name: "response with thermostats and no page info",
			opts: testServerOpts{
				Payload: `{
		"thermostatList": [
			{ "name": "thermostat1" },
			{ "name": "thermostat2" }
		],
		"status": { "code": 200, "message": "OK" }
	}`,
			},
			want: []*Thermostat{
				&Thermostat{Name: "thermostat1"},
				&Thermostat{Name: "thermostat2"},
			},
		},
		{
			name: "OK response with empty thermostat list",
			opts: testServerOpts{
				Payload: `{
		"page": {
			"page": 1,
			"totalPages": 1,
			"pageSize": 2,
			"total": 2
		},
		"thermostatList": [],
		"status": { "code": 200, "message": "OK" }
	}`,
			},
			want: []*Thermostat{},
		},
		{
			name: "OK response with no thermostat list",
			opts: testServerOpts{
				Payload: `{
		"page": {
			"page": 1,
			"totalPages": 1,
			"pageSize": 2,
			"total": 2
		},
		"status": { "code": 200, "message": "OK" }
	}`,
			},
		},
		{
			name: "not-ok (503) response",
			opts: testServerOpts{
				StatusCode: 503,
			},
			wantErr: "non-ok status response from API: 503 Internal Server Error",
		},
	} {
		opts := baseOpts
		if err := mergo.Merge(&opts, tt.opts, mergo.WithOverride); err != nil {
			t.Fatalf("failed setting up opts for test: %v", err)
		}
		client, server := clientAndServerForTest(t, opts)
		got, err := client.Thermostats(&Selection{})
		if tt.wantErr != "" && err == nil {
			t.Errorf("case %q: didn't get expected error; want: %q", tt.name, tt.wantErr)
		}
		if tt.wantErr == "" && err != nil {
			t.Errorf("case %q: got unexpected error: %v", tt.name, err)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("case %q: return value check failed;\ngot: %#v\nwant: %#v", tt.name, got, tt.want)
		}
		server.Close()
	}
}

func TestClientThermostatsJSONDecodeError(t *testing.T) {
	origJSONDecode := jsonDecode
	jsonDecode = func(io.Reader, interface{}) error {
		return errors.New("test error")
	}
	defer func() { jsonDecode = origJSONDecode }()

	client, server := clientAndServerForTest(t, testServerOpts{APIPath: "/1/thermostat"})
	defer server.Close()
	got, err := client.Thermostats(&Selection{})

	if got != nil {
		t.Errorf("got unexpected return value; got: %+v, want: nil", got)
	}
	if err.Error() != "test error" {
		t.Errorf(`got unexpected error value; got: %v, want: "test error"`, err)
	}
}

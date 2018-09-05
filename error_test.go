package egobee

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestUnmarshalErrorResponse(t *testing.T) {
	// This JSON has stray whitespace which is preserved from the source docs.
	exampleJSON := `{
    "error": "invalid_client",
    "error_description": "Authentication error, invalid authentication method, lack of credentials, etc.",
    "error_uri": "https://tools.ietf.org/html/rfc6749#section-5.2"
}`
	want := &ErrorResponse{
		Error:       APIErrorInvalidClient,
		Description: "Authentication error, invalid authentication method, lack of credentials, etc.",
		URI:         "https://tools.ietf.org/html/rfc6749#section-5.2",
	}

	got := &ErrorResponse{}
	if err := json.Unmarshal([]byte(exampleJSON), got); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, wanted: %+v", got, want)
	}
}

func TestErrorResponse_ParseString(t *testing.T) {
	exampleJSON := `{
    "error": "invalid_client",
    "error_description": "Authentication error, invalid authentication method, lack of credentials, etc.",
    "error_uri": "https://tools.ietf.org/html/rfc6749#section-5.2"
}`

	want := &ErrorResponse{
		Error:       APIErrorInvalidClient,
		Description: "Authentication error, invalid authentication method, lack of credentials, etc.",
		URI:         "https://tools.ietf.org/html/rfc6749#section-5.2",
	}

	got := &ErrorResponse{}
	if err := got.ParseString(exampleJSON); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, wanted: %+v", got, want)
	}
}

func TestErrorResponse_Parse(t *testing.T) {
	exampleJSON := []byte(`{
    "error": "invalid_client",
    "error_description": "Authentication error, invalid authentication method, lack of credentials, etc.",
    "error_uri": "https://tools.ietf.org/html/rfc6749#section-5.2"
}`)

	want := &ErrorResponse{
		Error:       APIErrorInvalidClient,
		Description: "Authentication error, invalid authentication method, lack of credentials, etc.",
		URI:         "https://tools.ietf.org/html/rfc6749#section-5.2",
	}

	got := &ErrorResponse{}
	if err := got.Parse(exampleJSON); err != nil {
		t.Errorf("got unexpected error: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got: %+v, wanted: %+v", got, want)
	}
}

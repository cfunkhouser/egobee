package egobee

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const ecobeeThermostatSumaryURL = "https://api.ecobee.com/1/thermostatSummary"

// summarySelection wraps a Selection, and serializes to the format expected by
// the thermostatSummary API.
type summarySelection struct {
	Selection Selection `json:"selection"`
}

// ThermostatSummary retrieves a list of thermostat configuration and state
// revisions. This API request is a light-weight polling method which will only
// return the revision numbers for the significant portions of the thermostat
// data.
// See https://www.ecobee.com/home/developer/api/documentation/v1/operations/get-thermostat-summary.shtml
func (c *Client) ThermostatSummary() (*ThermostatSummary, error) {
	s := &summarySelection{
		Selection: Selection{
			SelectionType: SelectionTypeRegistered,
			IncludeAlerts: true,
		},
	}
	qb, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	// The thermostatSummary API is unusual in that it requires JSON data, but
	// expects it to be attached as an encoded query parameter sent via GET,
	// instead of as the request body via POST as most ecobee API calls.
	url := fmt.Sprintf(`%v?json=%v`, ecobeeThermostatSumaryURL, url.QueryEscape(string(qb)))
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	r.Header.Add("Content-Type", "application/json; charset=utf-8")
	res, err := c.Do(r)
	if err != nil {
		log.Fatalf("Failed to Do request: %v", err)
	}
	defer res.Body.Close()
	ts := &ThermostatSummary{}
	if err := json.NewDecoder(res.Body).Decode(ts); err != nil {
		return nil, err
	}
	return ts, nil
}

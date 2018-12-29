package egobee

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

const (
	ecobeeThermostatSummaryURL = "https://api.ecobee.com/1/thermostatSummary"
	ecobeeThermostatURL        = "https://api.ecobee.com/1/thermostat"
)

// page is used for paging in some APIs.
type page struct {
	Page       int `json:"page"`
	TotalPages int `json:"totalPages"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
}

// summarySelection wraps a Selection, and serializes to the format expected by
// the thermostatSummary API.
type summarySelection struct {
	Selection Selection `json:"selection,omitempty"`
}

func assembleSelectURL(apiURL string, selection *Selection) (string, error) {
	ss := &summarySelection{
		Selection: *selection,
	}
	qb, err := json.Marshal(ss)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`%v?json=%v`, apiURL, url.QueryEscape(string(qb))), nil
}

// ThermostatSummary retrieves a list of thermostat configuration and state
// revisions. This API request is a light-weight polling method which will only
// return the revision numbers for the significant portions of the thermostat
// data.
// See https://www.ecobee.com/home/developer/api/documentation/v1/operations/get-thermostat-summary.shtml
func (c *Client) ThermostatSummary() (*ThermostatSummary, error) {
	url, err := assembleSelectURL(ecobeeThermostatSummaryURL, &Selection{
		SelectionType: SelectionTypeRegistered,
		IncludeAlerts: true,
	})
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	r.Header.Add("Content-Type", "application/json; charset=utf-8")
	res, err := c.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to Do(): %v", err)
	}
	defer res.Body.Close()
	if (res.StatusCode / 100) != 2 {
		return nil, fmt.Errorf("non-ok status response from API: %v", res.Status)
	}
	ts := &ThermostatSummary{}
	if err := json.NewDecoder(res.Body).Decode(ts); err != nil {
		return nil, err
	}
	return ts, nil
}

// See https://www.ecobee.com/home/developer/api/documentation/v1/operations/get-thermostats.shtml
type pagedThermostatResponse struct {
	Page        page          `json:"page,omitempty"`
	Thermostats []*Thermostat `json:"thermostatList,omitempty"`
	Status      struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
	} `json:"status,omitempty"`
}

// Thermostats returns all Thermostat objects which match selection.
func (c *Client) Thermostats(selection *Selection) ([]*Thermostat, error) {
	url, err := assembleSelectURL(ecobeeThermostatURL, selection)
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	r.Header.Add("Content-Type", "application/json; charset=utf-8")
	res, err := c.Do(r)
	if err != nil {
		return nil, fmt.Errorf("failed to Do(): %v", err)
	}
	defer res.Body.Close()
	if (res.StatusCode / 100) != 2 {
		return nil, fmt.Errorf("non-ok status response from API: %v", res.Status)
	}
	ptr := &pagedThermostatResponse{}
	if err := json.NewDecoder(res.Body).Decode(ptr); err != nil {
		return nil, err
	}
	if ptr.Page.Page != ptr.Page.TotalPages {
		// TODO(cfunkhouser): Handle paged responses.
		log.Printf("WARNING: Skipped some paged responses!")
	}
	return ptr.Thermostats, nil
}

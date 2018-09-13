// listthermostats uses the ThermostatSummary API call to list thermostats for
// an account.
package main

import (
	"flag"
	"log"
	"time"

	"github.com/cfunkhouser/egobee"
)

var accessToken = flag.String("access_token", "", "Ecobee API Access Token")

func main() {
	flag.Parse()
	if *accessToken == "" {
		log.Fatal("--access_token is require.")
	}

	ts := egobee.NewMemoryTokenStore(&egobee.TokenRefreshResponse{
		AccessToken: *accessToken,
		// Some non-zero value is all it should take.
		ExpiresIn: egobee.TokenDuration{Duration: time.Minute * 5},
	})
	c := egobee.New("94fl7gSO9SFmTXBKr0aSjzMwkjXAIRnZ", ts)

	summary, err := c.ThermostatSummary()
	if err != nil {
		log.Fatalf("This is no good: %+v", err)
	}

	for _, r := range summary.RevisionList {
		log.Println(r)
	}
}

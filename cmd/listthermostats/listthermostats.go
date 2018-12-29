// listthermostats uses the Thermostats API call to list the latest temperature
// readings for all registered thermostats.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/cfunkhouser/egobee"
)

var (
	accessToken = flag.String("access_token", "", "Ecobee API Access Token")
	appID       = flag.String("app_id", "", "Ecobee Registered App ID")
)

func main() {
	flag.Parse()
	if *accessToken == "" {
		log.Fatal("--access_token is required.")
	}
	if *appID == "" {
		log.Fatal("--app_id is required.")
	}

	ts := egobee.NewMemoryTokenStore(&egobee.TokenRefreshResponse{
		AccessToken: *accessToken,
		// Some non-zero value is all it should take.
		ExpiresIn: egobee.TokenDuration{Duration: time.Minute * 5},
	})
	c := egobee.New(*appID, ts)

	thermostats, err := c.Thermostats(&egobee.Selection{
		SelectionType:   egobee.SelectionTypeRegistered,
		IncludeSettings: true,
		IncludeRuntime:  true,
		IncludeSensors:  true,
	})
	if err != nil {
		log.Fatalf("This is no good: %+v", err)
	}
	for _, thermostat := range thermostats {
		fmt.Printf("%v currently averaging %v\n", thermostat.Name, float64(float64(thermostat.Runtime.ActualTemperature)/10))
		if len(thermostat.RemoteSensors) > 0 {
			for _, sensor := range thermostat.RemoteSensors {
				t, err := sensor.Temperature()
				if err != nil {
					log.Printf("Error getting temperature: %v", err)
					continue // Skip the bad sensor.
				}
				fmt.Printf("  %v currently %v\n", sensor.Name, t)
			}
		}
	}
}

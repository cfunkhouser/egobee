// listthermostats uses the Thermostats API call to list the latest temperature
// readings for all registered thermostats.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/cfunkhouser/egobee"
)

var (
	appID     = flag.String("app", "", "Ecobee Registered App ID")
	storePath = flag.String("store", "", "Persistent egobee credential store path")
)

func main() {
	flag.Parse()

	if *appID == "" {
		log.Fatal("--app is required")
	}
	if *storePath == "" {
		log.Fatal("--store is required")
	}

	ts, err := egobee.NewPersistentTokenStorer(*storePath)
	if err != nil {
		log.Fatalf("Failed to read store %q: %v", *storePath, err)
	}
	c := egobee.New(*appID, ts)
	if !ts.Initialized() {
		log.Fatalf("Store %q not initialized", *storePath)
	}

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

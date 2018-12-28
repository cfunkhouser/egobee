package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/cfunkhouser/egobee"
)

const (
	authURLTemplate = `https://api.ecobee.com/authorize?response_type=ecobeePin&scope=smartWrite&client_id=%v`
	tokenURL        = "https://api.ecobee.com/token"
)

var (
	appID     = flag.String("app", "", "Ecobee Registered App ID")
	storePath = flag.String("store", "/tmp/promobee", "Persistent egobee credential store path")
)

func main() {
	flag.Parse()
	if *appID == "" {
		log.Fatal("--app is required.")
	}
	if *storePath == "" {
		log.Fatal("--store is required")
	}

	ts, err := egobee.NewPersistentTokenStorer(*storePath)
	if err != nil {
		log.Fatalf("Couldn't use store at %q: %v", *storePath, err)
	}
	auth := egobee.NewPinAuthenticator(*appID)
	pin, err := auth.GetPin()
	if err != nil {
		log.Fatalf("Failed to initialize app %v: %v", *appID, err)
	}

	// Up to the caller to complete the auth flow out-of-band.
	fmt.Printf("Register with PIN: %v\n", pin)
	fmt.Println("Press any key to continue when done.")

	var input string
	fmt.Scanf("%s", &input)

	if err := auth.Finalize(ts); err != nil {
		log.Fatalf("Failed to initalize app %v: %v", *appID, err)
	}
	fmt.Printf("Created persistent store at %q\n", *storePath)
}

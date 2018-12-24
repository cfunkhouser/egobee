package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

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

	resp, err := http.Get(fmt.Sprintf(authURLTemplate, *appID))
	if err != nil {
		log.Fatalf("Failed to initialize Pin Authentication: %v", err)
	}

	pac := &egobee.PinAuthenticationChallenge{}
	if err := json.NewDecoder(resp.Body).Decode(pac); err != nil {
		log.Fatalf("Failed to read Pin Authentication: %v", err)
	}
	resp.Body.Close()

	fmt.Printf("Register with this PIN: %v\n", pac.Pin)
	fmt.Println("Press any key to continue when done.")

	var input string
	fmt.Scanf("%s", &input)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "grant_type=ecobeePin&code=%v&client_id=%v", pac.AuthorizationCode, *appID)

	resp, err = http.Post(tokenURL, "application/x-www-form-urlencoded", &buf)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v")
	}
	defer resp.Body.Close()

	trr := &egobee.TokenRefreshResponse{}
	if err := json.NewDecoder(resp.Body).Decode(trr); err != nil {
		log.Fatalf("Failed to decode authentication response: %v", err)
	}
	_, err = egobee.NewPersistentTokenStore(trr, *storePath)
	if err != nil {
		log.Fatalf("Failed to initialize persistent store: %v", err)
	}
	fmt.Printf("Created persistent store at %v\n", *storePath)
}

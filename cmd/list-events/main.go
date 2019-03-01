package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func main() {
	apiKey := os.Getenv("LINODE_TOKEN")

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	oc := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linode := linodego.NewClient(oc)
	events, err := linode.ListEvents(context.Background(), linodego.NewListOptions(1, ""))
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		fmt.Println("---")
		fmt.Printf("ID: %d\n", event.ID)
		fmt.Printf("Action: %v\n", event.Action)
		fmt.Printf("Username: %v\n", event.Username)
		fmt.Printf("Entity: ")
		if event.Entity == nil {
			fmt.Printf("%v\n", event.Entity)
		} else {
			fmt.Printf("\n  ID: %.f\n", event.Entity.ID)
			fmt.Printf("  Label: %v\n", event.Entity.Label)
			fmt.Printf("  Type: %v\n", event.Entity.Type)
			fmt.Printf("  URL: %v\n", event.Entity.URL)
		}
		fmt.Printf("Read: %v\n", event.Read)
		fmt.Printf("Seen: %v\n", event.Seen)
		fmt.Printf("Status: %v\n", event.Status)
		fmt.Printf("Created: %v\n", event.Created)
		fmt.Printf("TimeRemaining: %s\n", event.TimeRemainingMsg)
		fmt.Printf("Rate: %v\n", event.Rate)
	}
	fmt.Println("---")

}

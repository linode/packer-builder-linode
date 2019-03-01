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
	images, err := linode.ListImages(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	for _, image := range images {
		fmt.Printf("%+v\n", image)
	}
}

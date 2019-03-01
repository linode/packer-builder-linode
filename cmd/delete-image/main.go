package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func main() {
	apiKey := os.Getenv("LINODE_TOKEN")
	imageID, err := strconv.Atoi(os.Args[1])

	if err != nil {
		panic(err)
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	oc := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}

	linode := linodego.NewClient(oc)
	err = linode.DeleteImage(context.Background(), fmt.Sprintf("private/%d", imageID))
	if err != nil {
		panic(err)
	}

	fmt.Println("deleted image", imageID)
}

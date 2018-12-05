package linode

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform/helper/logging"
	"github.com/hashicorp/terraform/version"
	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func newLinodeClient(pat string) linodego.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: pat})

	oauthTransport := &oauth2.Transport{
		Source: tokenSource,
	}
	loggingTransport := logging.NewTransport("Linode", oauthTransport)
	oauth2Client := &http.Client{
		Transport: loggingTransport,
	}

	client := linodego.NewClient(oauth2Client)

	projectURL := "https://www.packer.io"
	userAgent := fmt.Sprintf("Packer/%s (+%s) linodego/%s",
		version.String(), projectURL, linodego.Version)

	client.SetUserAgent(userAgent)
	return client
}

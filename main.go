package main

import (
	"github.com/linode/packer-builder-linode/linode"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(linode.Builder))
	server.Serve()
}

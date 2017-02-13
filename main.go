package main

import (
	"github.com/dradtke/packer-builder-linode/linode"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(linode.Builder))
	server.Serve()
}

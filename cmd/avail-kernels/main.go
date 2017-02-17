package main

import (
	"context"
	"fmt"
	"os"

	// "github.com/dradtke/packer-builder-linode/linode"
	"github.com/mitchellh/packer/builder/linode"
)

func main() {
	apiKey := os.Getenv("LINODE_API_KEY")
	kernels, err := linode.AvailKernels(context.Background(), apiKey)
	if err != nil {
		panic(err)
	}
	for _, kernel := range kernels {
		fmt.Println(kernel)
	}
}

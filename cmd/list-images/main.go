package main

import (
	"context"
	"fmt"
	"os"

	"github.com/dradtke/packer-builder-linode/linode"
)

func main() {
	apiKey := os.Getenv("LINODE_API_KEY")
	images, err := linode.ImageList(context.Background(), apiKey, false, 0)
	if err != nil {
		panic(err)
	}
	fmt.Println(images)
}

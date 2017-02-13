package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/dradtke/packer-builder-linode/linode"
)

func main() {
	apiKey := os.Getenv("LINODE_API_KEY")
	imageId, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}

	if err := linode.ImageDelete(context.Background(), apiKey, imageId); err != nil {
		panic(err)
	}
	fmt.Println("deleted image", imageId)
}

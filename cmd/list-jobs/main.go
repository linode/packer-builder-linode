package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	// "github.com/dradtke/packer-builder-linode/linode"
	"github.com/mitchellh/packer/builder/linode"
)

var (
	jobId    = flag.Int("job", 0, "job id")
	linodeId = flag.Int("linode", 0, "linode id")
)

func main() {
	flag.Parse()
	if *linodeId == 0 {
		fmt.Println("No linode id specified.")
		os.Exit(1)
	}

	apiKey := os.Getenv("LINODE_API_KEY")
	jobs, err := linode.LinodeJobList(context.Background(), apiKey, *linodeId, *jobId, false)
	if err != nil {
		panic(err)
	}
	for _, job := range jobs {
		fmt.Println("---------------------------------")
		fmt.Printf("ID: %v\n", job.ID)
		fmt.Printf("LinodeID: %v\n", job.LinodeID)
		fmt.Printf("EnteredDate: %v\n", job.EnteredDate)
		fmt.Printf("HostStartDate: %v\n", job.HostStartDate)
		fmt.Printf("HostFinishDate: %v\n", job.HostFinishDate)
		fmt.Printf("Action: %v\n", job.Action)
		fmt.Printf("Label: %v\n", job.Label)
		fmt.Printf("Duration: %v\n", job.Duration)
		fmt.Printf("HostMessage: %v\n", job.HostMessage)
		fmt.Printf("HostSuccess: %v\n", job.HostSuccess)
	}
	fmt.Println("---------------------------------")
}

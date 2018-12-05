package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/linode/linodego"
)

type stepCreateImage struct {
	client linodego.Client
}

func (s *stepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	disk := state.Get("disk").(*linodego.InstanceDisk)

	ui.Say("Creating image...")
	image, err := s.client.CreateImage(ctx, linodego.ImageCreateOptions{
		DiskID:      disk.ID,
		Label:       c.Label,
		Description: c.Description,
	})
	if err != nil {
		err = errors.New("Error creating image: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// TODO: We need to find some way to wait until the image is done being created.
	// With this implementation as it stands, the image gets cleaned up almost immediately for some reason.
	// My best guess is that it gets deleted when we clean up the Linode instance because it's still
	// being used in the imaging process.

	state.Put("image", image)
	return multistep.ActionContinue
}

func (s *stepCreateImage) Cleanup(state multistep.StateBag) {}

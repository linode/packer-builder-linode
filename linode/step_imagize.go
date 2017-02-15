package linode

import (
	"context"
	"errors"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepImagize struct{}

func (s *stepImagize) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Imagizing...")
	imageId, jobId, err := LinodeDiskImagize(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
		state.Get("disk_id").(int),
		c.Description,
		c.Label,
	)
	if err != nil {
		err = errors.New("Error creating image: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := waitForJob(ui, ctx, c, state.Get("linode_id").(int), jobId); err != nil {
		err = errors.New("Error creating image: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("image_id", imageId)
	return multistep.ActionContinue
}

func (s *stepImagize) Cleanup(state multistep.StateBag) {}

package linode

import (
	"context"
	"errors"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateDisk struct{}

func (s *stepCreateDisk) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating disk...")
	diskId, jobId, err := LinodeDiskCreateFromDistribution(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
		c.DistributionID,
		c.Label,
		c.DiskSize,
		c.RootPass,
		c.RootSSHKey,
	)
	if err != nil {
		err = errors.New("Error creating disk: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := waitForJob(ui, ctx, c, state.Get("linode_id").(int), jobId); err != nil {
		err = errors.New("Error creating disk: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("disk_id", diskId)
	return multistep.ActionContinue
}

func (s *stepCreateDisk) Cleanup(state multistep.StateBag) {
	diskId, ok := state.GetOk("disk_id")
	if !ok {
		return
	}

	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Cleaning up...")
	jobId, err := LinodeDiskDelete(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
		diskId.(int),
	)
	if err != nil {
		ui.Error("Error cleaning up Linode disk: " + err.Error())
		return
	}

	if err := waitForJob(ui, ctx, c, state.Get("linode_id").(int), jobId); err != nil {
		ui.Error("Error cleaning up Linode disk: " + err.Error())
	}
}

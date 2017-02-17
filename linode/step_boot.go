package linode

import (
	"context"
	"errors"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepBoot struct {
	booted bool
}

func (s *stepBoot) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Booting Linode...")
	jobId, err := LinodeBoot(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
		state.Get("config_id").(int),
	)
	if err != nil {
		err = errors.New("Error booting Linode: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err := waitForJob(ui, ctx, c, state.Get("linode_id").(int), jobId); err != nil {
		err = errors.New("Error booting Linode: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Ensure that it's actually running before continuing
	ui.Message("Waiting for Running status...")
	for {
		linodes, err := LinodeList(ctx, c.APIKey, state.Get("linode_id").(int))
		if err != nil {
			err = errors.New("Error getting Linode status: " + err.Error())
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		if linodes[0].Running() {
			break
		}
		time.Sleep(2 * time.Second)
	}

	s.booted = true
	return multistep.ActionContinue
}

func (s *stepBoot) Cleanup(state multistep.StateBag) {
	if !s.booted {
		return
	}

	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	jobId, err := LinodeShutdown(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
	)
	if err != nil {
		ui.Error("Error shutting down Linode: " + err.Error())
	}

	if err := waitForJob(ui, ctx, c, state.Get("linode_id").(int), jobId); err != nil {
		ui.Error("Error shutting down Linode: " + err.Error())
	}
}

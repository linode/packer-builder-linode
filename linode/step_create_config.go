package linode

import (
	"context"
	"errors"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepCreateConfig struct{}

func (s *stepCreateConfig) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating boot config...")
	configId, err := LinodeConfigCreate(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
		state.Get("disk_id").(int),
		c.KernelID,
		c.Label,
	)
	if err != nil {
		err = errors.New("Error creating boot config: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("config_id", configId)
	return multistep.ActionContinue
}

func (s *stepCreateConfig) Cleanup(state multistep.StateBag) {
	configId, ok := state.GetOk("config_id")
	if !ok {
		return
	}

	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	err := LinodeConfigDelete(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
		configId.(int),
	)
	if err != nil {
		ui.Error("Error cleaning up Linode config: " + err.Error())
		return
	}
}

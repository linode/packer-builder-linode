package linode

import (
	"context"
	"errors"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepLinodeIP struct{}

func (s *stepLinodeIP) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	ips, err := LinodeIPList(
		ctx,
		c.APIKey,
		state.Get("linode_id").(int),
		0,
	)
	if err != nil {
		err = errors.New("Error retrieving Linode IP: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("linode_ip", ips[0])
	return multistep.ActionContinue
}

func (s *stepLinodeIP) Cleanup(state multistep.StateBag) {}

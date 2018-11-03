package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepLinodeIP struct{}

func (s *stepLinodeIP) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
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

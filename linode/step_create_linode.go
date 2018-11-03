package linode

import (
	"context"
	"errors"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateLinode struct{}

func (s *stepCreateLinode) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating Linode...")
	linodeId, err := LinodeCreate(
		ctx,
		c.APIKey,
		c.DatacenterID,
		c.PlanID,
		c.PaymentTerm,
	)
	if err != nil {
		err = errors.New("Error creating Linode: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("linode_id", linodeId)
	return multistep.ActionContinue
}

func (s *stepCreateLinode) Cleanup(state multistep.StateBag) {
	linodeId, ok := state.GetOk("linode_id")
	if !ok {
		return
	}

	c := state.Get("config").(Config)
	ctx := state.Get("ctx").(context.Context)
	ui := state.Get("ui").(packer.Ui)

	if err := LinodeDelete(
		ctx,
		c.APIKey,
		linodeId.(int),
		false,
	); err != nil {
		ui.Error("Error cleaning up Linode: " + err.Error())
	}
}

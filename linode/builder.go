package linode

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
)

type Builder struct {
	config Config

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
}

func (b *Builder) Prepare(raws ...interface{}) (warnings []string, err error) {
	if err = config.Decode(&b.config, &config.DecodeOpts{
		Interpolate: true,
	}, raws...); err != nil {
		return warnings, err
	}

	b.ctx, b.cancel = context.WithCancel(context.Background())
	b.done = make(chan struct{})

	if b.config.APIKey == "" {
		return warnings, errors.New("configuration value `api_key` not defined")
	}

	/* -- Distribution -- */

	var distros []Distribution
	if b.config.DistributionID == 0 {
		if distros, err = AvailDistributions(b.ctx, b.config.APIKey, 0); err != nil {
			return warnings, err
		}
	}
	showDistros := func() {
		if distros == nil {
			return
		}
		log.Print("Available distributions:")
		for _, distro := range distros {
			log.Print("--------------------------")
			log.Printf("  ID: %d", distro.ID)
			log.Printf("  Label: %s", distro.Label)
		}
	}
	if b.config.DistributionID == 0 && b.config.DistributionName != "" {
		for _, distro := range distros {
			if distro.Label == b.config.DistributionName {
				b.config.DistributionID = distro.ID
				break
			}
		}
		if b.config.DistributionID == 0 {
			showDistros()
			return warnings, fmt.Errorf("no distribution found with label \"%s\" (run again with PACKER_LOG set to see available distributions)", b.config.DistributionName)
		}
	}

	if b.config.DistributionID == 0 {
		showDistros()
		return warnings, errors.New("one configuration value of `distribution_name` or `distribution_id` must be specified (run again with PACKER_LOG set to see available distributions)")
	}

	/* -- Datacenter -- */

	var datacenters []Datacenter
	if b.config.DatacenterID == 0 {
		if datacenters, err = AvailDatacenters(b.ctx, b.config.APIKey); err != nil {
			return warnings, err
		}
	}
	showDatacenters := func() {
		if datacenters == nil {
			return
		}
		log.Print("Available datacenters:")
		for _, dc := range datacenters {
			log.Print("--------------------------")
			log.Printf("  ID: %d", dc.ID)
			log.Printf("  Location: %s", dc.Location)
			log.Printf("  Abbr: %s", dc.Abbr)
		}
	}
	if b.config.DatacenterID == 0 && b.config.DatacenterName != "" {
		for _, dc := range datacenters {
			if dc.Location == b.config.DatacenterName || dc.Abbr == b.config.DatacenterName {
				b.config.DatacenterID = dc.ID
				break
			}
		}
		if b.config.DatacenterID == 0 {
			showDatacenters()
			return warnings, fmt.Errorf("no datacenter found with name or abbreviation \"%s\" (run again with PACKER_LOG set to see available datacenters)", b.config.DatacenterName)
		}
	}

	if b.config.DatacenterID == 0 {
		showDatacenters()
		return warnings, errors.New("one configuration value of `datacenter_name` or `datacenter_id` must be specified (run again with PACKER_LOG set to see available datacenters)")
	}

	/* -- Plan -- */

	var plans []Plan
	if b.config.PlanID == 0 {
		if plans, err = AvailPlans(b.ctx, b.config.APIKey, 0); err != nil {
			return warnings, err
		}
	}
	showPlans := func() {
		if plans == nil {
			return
		}
		log.Print("Available plans:")
		for _, plan := range plans {
			log.Print("--------------------------")
			log.Printf("  ID: %d", plan.ID)
			log.Printf("  Label: %s", plan.Label)
			log.Printf("  Cores: %d", plan.Cores)
			log.Printf("  RAM: %d", plan.RAM)
			log.Printf("  Xfer: %d", plan.Xfer)
			log.Printf("  Price: %f", plan.Price)
			log.Printf(fmt.Sprintf("  %d: %s", plan.ID, plan.Label))
		}
	}
	if b.config.PlanID == 0 && b.config.PlanName != "" {
		for _, plan := range plans {
			if plan.Label == b.config.PlanName {
				b.config.PlanID = plan.ID
				break
			}
		}
		if b.config.PlanID == 0 {
			showPlans()
			return warnings, fmt.Errorf("no plan found with name \"%s\" (run again with PACKER_LOG set to see available plans)", b.config.PlanName)
		}
	}

	if b.config.PlanID == 0 {
		showPlans()
		return warnings, errors.New("one configuration value of `plan_name` or `plan_id` must be specified (run again with PACKER_LOG set to see available plans)")
	}

	if b.config.Label == "" {
		return warnings, errors.New("configuration value `label` not defined")
	}

	if b.config.DiskSize == 0 {
		return warnings, errors.New("configuration value `disk_size` not defined")
	}

	if b.config.RootPass == "" {
		return warnings, errors.New("configuration value `root_pass` not defined")
	}

	return warnings, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (ret packer.Artifact, err error) {
	defer close(b.done)

	ui.Say("Creating new Linode")
	artifact := Artifact{apiKey: b.config.APIKey}
	defer func() {
		if err != nil && artifact.LinodeID != 0 {
			ui.Say("An error was encountered, deleting the Linode")
			if err2 := LinodeDelete(b.ctx, artifact.apiKey, artifact.LinodeID, false); err2 != nil {
				err = multierror.Append(err, err2)
			}
		}
	}()
	if artifact.LinodeID, err = LinodeCreate(b.ctx, b.config.APIKey, b.config.DatacenterID, b.config.PlanID, b.config.PaymentTerm); err != nil {
		return nil, err
	}

	ui.Say("Creating disk")
	if artifact.DiskID, err = LinodeDiskCreateFromDistribution(b.ctx, b.config.APIKey, artifact.LinodeID, b.config.DistributionID, b.config.Label, b.config.DiskSize, b.config.RootPass, b.config.RootSSHKey); err != nil {
		return nil, err
	}

	// TODO: run provisioners

	ui.Say("Imagizing the disk")
	if artifact.ImageID, err = LinodeDiskImagize(b.ctx, b.config.APIKey, artifact.LinodeID, artifact.DiskID, b.config.Description, b.config.Label); err != nil {
		return nil, err
	}

	// ui.Say("Done, cleaning up")
	return artifact, nil
}

func (b *Builder) Cancel() {
	b.cancel()
	<-b.done
}

package linode

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

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
	var (
		linodeId int
		diskId   int
		imageId  int
		jobId    int
		ips      []LinodeIP
	)

	ui.Say("Creating new Linode")
	defer func() {
		if err != nil && linodeId != 0 {
			ui.Say("An error was encountered, deleting the Linode")
			if diskId != 0 {
				var err2 error
				if jobId, err2 = LinodeDiskDelete(b.ctx, b.config.APIKey, linodeId, diskId); err2 != nil {
					err = multierror.Append(err, err2)
				}
				if err2 = b.waitForJob(ui, linodeId, jobId); err != nil {
					err = multierror.Append(err, err2)
				}
			}
			if err2 := LinodeDelete(b.ctx, b.config.APIKey, linodeId, false); err2 != nil {
				err = multierror.Append(err, err2)
			}
		}
	}()
	if linodeId, err = LinodeCreate(b.ctx, b.config.APIKey, b.config.DatacenterID, b.config.PlanID, b.config.PaymentTerm); err != nil {
		return nil, err
	}
	ui.Message(fmt.Sprintf("Linode ID: %d", linodeId))

	ui.Say("Creating disk")
	if diskId, jobId, err = LinodeDiskCreateFromDistribution(b.ctx, b.config.APIKey, linodeId, b.config.DistributionID, b.config.Label, b.config.DiskSize, b.config.RootPass, b.config.RootSSHKey); err != nil {
		return nil, err
	}
	if err = b.waitForJob(ui, linodeId, jobId); err != nil {
		return nil, err
	}
	ui.Message(fmt.Sprintf("Disk ID: %d", diskId))

	if ips, err = LinodeIPList(b.ctx, b.config.APIKey, linodeId, 0); err != nil {
		return nil, err
	}
	ui.Message(fmt.Sprintf("Linode IP Address: %s", ips[0].Address))

	// TODO: run provisioners

	ui.Say("Imagizing the disk")
	if imageId, jobId, err = LinodeDiskImagize(b.ctx, b.config.APIKey, linodeId, diskId, b.config.Description, b.config.Label); err != nil {
		return nil, err
	}
	if err = b.waitForJob(ui, linodeId, jobId); err != nil {
		return nil, err
	}

	cleanup := func() error {
		ui.Say("Cleaning up")
		if jobId, err = LinodeDiskDelete(b.ctx, b.config.APIKey, linodeId, diskId); err != nil {
			return errors.New("failed to start disk deletion: " + err.Error())
		}
		if err = b.waitForJob(ui, linodeId, jobId); err != nil {
			return errors.New("disk deletion failed: " + err.Error())
		}
		if err = LinodeDelete(b.ctx, b.config.APIKey, linodeId, false); err != nil {
			return errors.New("failed to delete Linode: " + err.Error())
		}
		return nil
	}

	if err = cleanup(); err != nil {
		ui.Say("Image was built successfully, but an error was encountered during cleanup: " + err.Error())
		ui.Say("You should delete Linode " + strconv.Itoa(linodeId) + " manually.")
	}

	// ui.Say("Done, cleaning up")
	return Artifact{apiKey: b.config.APIKey, ImageID: imageId}, nil
}

func (b *Builder) waitForJob(ui packer.Ui, linodeId, jobId int) error {
	ui.Message("--> Waiting for job " + strconv.Itoa(jobId) + " to complete")
	for {
		jobs, err := LinodeJobList(b.ctx, b.config.APIKey, linodeId, jobId, false)
		if err != nil {
			return err
		}
		if jobs[0].HostFinishDate != "" {
			if int(jobs[0].HostSuccess) == 0 {
				return errors.New(jobs[0].HostMessage)
			}
			break
		}
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (b *Builder) Cancel() {
	b.cancel()
	<-b.done
}

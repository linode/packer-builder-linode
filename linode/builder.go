// The linode package contains a packer.Builder implementation
// that builds Linode images.
package linode

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hashicorp/packer/common"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// The unique ID for this builder.
const BuilderID = "packer.linode"

// Builder represents a Packer Builder.
type Builder struct {
	config Config
	runner multistep.Runner

	ctxCancel context.CancelFunc
	cancel    context.CancelFunc
}

func (b *Builder) Prepare(raws ...interface{}) (warnings []string, err error) {
	c, errs := NewConfig(raws...)
	if errs != nil {
		return nil, errs
	}
	b.config = c

	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (ret packer.Artifact, err error) {
	ui.Say("Running builder ...")

	ctx, cancel := context.WithCancel(context.Background())
	b.ctxCancel = cancel
	defer cancel()

	client := newLinodeClient(b.Config.Token)

	if err != nil {
		ui.Error(err.Error())
		return nil, err
	}

	state := new(multistep.BasicStateBag)
	state.Put("client", client)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)

	steps := []multistep.Step{
		&StepCreateSSHKey{
			Debug:          b.config.PackerDebug,
			DebugKeyPath:   fmt.Sprintf("linode_%s.pem", b.config.PackerBuildName),
			PrivateKeyFile: b.config.Comm.SSHPrivateKey,
		},
		new(stepCreateLinode),
		new(stepCreateDisk),
		new(stepCreateConfig),
		new(stepBoot),
		new(stepLinodeIP),
		// new(stepConnect),
		// new(stepProvision),
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		new(common.StepProvision),
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		new(stepImagize),
	}

	b.runner = common.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(state)

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	if _, ok := state.GetOk("image_label"); !ok {
		return nil, errors.New("Cannot find image_label in state.")
	}

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	artifact := Artifact{
		client:     client,
		ImageLabel: state.Get("image_label").(string),
		ImageID:    state.Get("image_id").(int),
	}

	return artifact, nil
}

func waitForJob(ui packer.Ui, ctx context.Context, config Config, linodeId, jobId int) error {
	ui.Message("--> Waiting for job " + strconv.Itoa(jobId) + " to complete")
	for {
		jobs, err := LinodeJobList(ctx, config.APIKey, linodeId, jobId, false)
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

// Cancel.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

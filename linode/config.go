package linode

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	Comm                communicator.Config `mapstructure:",squash"`

	PersonalAccessToken string `mapstructure:"personal_access_token"`

	Region string `mapstructure:"region"`

	InstanceType string `mapstructure:"instance_type"`

	Image string `mapstructure:"image"`

	DiskSize int    `mapstructure:"disk_size"`
	RootPass string `mapstructure:"root_pass"`

	// Optional label for the Linode Instance created
	Label string
	// Optional tags for the Linode Instance created
	Tags []string

	// Optional label for the Linode Image created
	ImageLabel string //optional

	// Optional Description for the Linode Image created
	Description string

	// Optional SSH Key for the root account
	RootSSHKey string

	RawStateTimeout string `mapstructure:"state_timeout"`

	stateTimeout time.Duration
	interCtx     interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

	if err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...); err != nil {
		return nil, nil, err
	}

	var errs *packer.MultiError

	// Defaults
	if c.ImageLabel == "" {
		if def, err := interpolate.Render("packer-{{timestamp}}", nil); err == nil {
			c.ImageLabel = def
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Unable to render image name: %s", err))
		}
	}

	if c.Label == "" {
		// Default to packer-[time-ordered-uuid]
		c.Label = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.RawStateTimeout == "" {
		c.stateTimeout = 5 * time.Minute
	} else {
		if stateTimeout, err := time.ParseDuration(c.RawStateTimeout); err == nil {
			c.stateTimeout = stateTimeout
		} else {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("Unable to parse state timeout: %s", err))
		}
	}

	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}

	if c.PersonalAccessToken == "" {
		// Required configurations that will display errors if not set
		errs = packer.MultiErrorAppend(
			errs, errors.New("personal access token is required"))
	}

	if c.Region == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("region is required"))
	}

	if c.InstanceType == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("instance_type is required"))
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(
			errs, errors.New("image is required"))
	}

	if c.Tags == nil {
		c.Tags = make([]string, 0)
	}
	tagRe := regexp.MustCompile("^[[:alnum:]:_-]{1,255}$")

	for _, t := range c.Tags {
		if !tagRe.MatchString(t) {
			errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("invalid tag: %s", t)))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.PersonalAccessToken)
	return c, nil, nil
}

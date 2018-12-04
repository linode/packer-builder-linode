package linode

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	Comm                communicator.Config `mapstructure:",squash"`

	Token  string `mapstructure:"token"`
	APIURL string `mapstructure:"api_url"`

	Region string `mapstructure:"region"`

	InstanceType string `mapstructure:"instance_type"`

	Image string `mapstructure:"image"`

	DiskSize int    `mapstructure:"disk_size"`
	RootPass string `mapstructure:"root_pass"`

	// Optional label for the Linode Instance created
	Label string
	// Optional tags for the Linode Instance created
	Tags string

	// Optional label for the Linode Image created
	ImageLabel string //optional

	// Optional Description for the Linode Image created
	Description string

	// Optional SSH Key for the root account
	RootSSHKey string

	interCtx interpolate.Context
}

func newConfig(raws ...interface{}) (c *Config, warnings []string, err error) {
	c = new(Config)

	var md mapstructure.Metadata
	err = config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"run_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Defaults
	if c.Token == "" {
		// Default to environment variable for api_token, if it exists
		c.Token = os.Getenv("LINODE_TOKEN")
	}
	if c.APIURL == "" {
		c.APIURL = os.Getenv("LINODE_API_URL")
	}
	if c.ImageLabel == "" {
		def, err := interpolate.Render("packer-{{timestamp}}", nil)
		if err != nil {
			panic(err)
		}

		// Default to packer-{{ unix timestamp (utc) }}
		c.ImageLabel = def
	}

	if c.Label == "" {
		// Default to packer-[time-ordered-uuid]
		c.Label = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}

	if c.StateTimeout == 0 {
		// Default to 6 minute timeouts waiting for
		// desired state. i.e waiting for droplet to become active
		c.StateTimeout = 6 * time.Minute
	}

	var errs *packer.MultiError
	if es := c.Comm.Prepare(&c.ctx); len(es) > 0 {
		errs = packer.MultiErrorAppend(errs, es...)
	}
	if c.Token == "" {
		// Required configurations that will display errors if not set
		errs = packer.MultiErrorAppend(
			errs, errors.New("token is required"))
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

	packer.LogSecretFilter.Set(c.Token)
	return c, nil, nil
}

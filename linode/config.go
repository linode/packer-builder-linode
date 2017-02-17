package linode

import (
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	ctx                 interpolate.Context
	Comm                communicator.Config `mapstructure:",squash"`

	APIKey string `mapstructure:"api_key"`

	DatacenterID   int    `mapstructure:"datacenter_id"`
	DatacenterName string `mapstructure:"datacenter_name"`

	PlanID   int    `mapstructure:"plan_id"`
	PlanName string `mapstructure:"plan_name"`

	DistributionID   int    `mapstructure:"distribution_id"`
	DistributionName string `mapstructure:"distribution_name"`

	KernelID   int    `mapstructure:"kernel_id"`
	KernelName string `mapstructure:"kernel_name"`

	DiskSize int    `mapstructure:"disk_size"`
	RootPass string `mapstructure:"root_pass"`

	Label string

	Description string // optional
	RootSSHKey  string // optional
	PaymentTerm int    // optional

	interCtx interpolate.Context
}

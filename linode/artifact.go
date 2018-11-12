package linode

import (
	"errors"
	"strconv"

	"github.com/linode/linodego"
)

type Artifact struct {
	ImageID    int
	ImageLabel string
	// The client for making API calls
	client linodego.Client
}

func (a Artifact) BuilderId() string             { return BuilderID }
func (a Artifact) Files() []string               { return nil }
func (a Artifact) Id() string                    { return strconv.Itoa(a.ImageID) }
func (a Artifact) String() string                { return "Linode image: " + a.Id() }
func (a Artifact) State(name string) interface{} { return nil }
func (a Artifact) Destroy() error {
	return errors.New("not implemented")
}

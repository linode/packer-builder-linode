package linode

import (
	"errors"
	"strconv"
)

type Artifact struct {
	ImageID int
	apiKey  string
}

func (a Artifact) BuilderId() string             { return "linode" }
func (a Artifact) Files() []string               { return nil }
func (a Artifact) Id() string                    { return strconv.Itoa(a.ImageID) }
func (a Artifact) String() string                { return "Linode image: " + a.Id() }
func (a Artifact) State(name string) interface{} { return nil }
func (a Artifact) Destroy() error {
	return errors.New("not implemented")
}

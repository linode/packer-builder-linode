package linode

import (
	"errors"
	"fmt"
)

type Artifact struct {
	ImageID    string
	ImageLabel string
}

func (a Artifact) BuilderId() string { return BuilderID }
func (a Artifact) Files() []string   { return nil }
func (a Artifact) Id() string        { return a.ImageID }
func (a Artifact) String() string {
	return fmt.Sprintf("Linode image: %s (%s)", a.ImageLabel, a.ImageID)
}
func (a Artifact) State(name string) interface{} { return nil }
func (a Artifact) Destroy() error {
	return errors.New("not implemented")
}

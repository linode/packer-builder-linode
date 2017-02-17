package linode

import (
	"errors"
	"io"
	"os"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"golang.org/x/crypto/ssh"
)

type stepProvision struct{}

func (s *stepProvision) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	comm := state.Get("communicator").(packer.Communicator)
	hook := state.Get("hook").(packer.Hook)

	ui.Say("Provisioning...")
	hook.Run(packer.HookProvision, ui, comm, nil)

	return multistep.ActionContinue
}

func (s *stepProvision) Cleanup(state multistep.StateBag) {}

type sshCommunicator struct {
	output *uiWriter
	client *ssh.Client
}

func (c *sshCommunicator) Start(cmd *packer.RemoteCmd) error {
	c.output.ui.Message("Running command: " + cmd.Command)
	session, err := c.client.NewSession()
	if err != nil {
		return errors.New("failed to start new SSH session: " + err.Error())
	}
	session.Stdout = c.output
	session.Stderr = c.output
	return session.Run(cmd.Command)
}

func (c *sshCommunicator) Upload(string, io.Reader, *os.FileInfo) error {
	panic("not implemented")
}

func (c *sshCommunicator) UploadDir(dst, src string, exclude []string) error {
	panic("not implemented")
}

func (c *sshCommunicator) Download(string, io.Writer) error {
	panic("not implemented")
}

func (c *sshCommunicator) DownloadDir(src, dst string, exclude []string) error {
	panic("not implemented")
}

type uiWriter struct {
	ui packer.Ui
}

func (w uiWriter) Write(b []byte) (int, error) {
	w.ui.Message(string(b))
	return len(b), nil
}

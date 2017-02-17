package linode

import (
	"errors"
	"strconv"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"golang.org/x/crypto/ssh"
)

type stepConnect struct{}

func (s *stepConnect) Run(state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(Config)
	ui := state.Get("ui").(packer.Ui)
	ip := state.Get("linode_ip").(LinodeIP).Address

	// Setting Timeout doesn't seem to work...
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password(c.RootPass)},
	}
	tries := 0
	ui.Say("Establishing connection @ " + ip + "...")
	for {
		if conn, err := ssh.Dial("tcp", ip+":22", config); err != nil {
			tries++
			if tries >= 5 {
				err = errors.New("Error establishing connection: " + err.Error())
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
			ui.Message("Attempt #" + strconv.Itoa(tries+1))
		} else {
			state.Put("communicator", &sshCommunicator{&uiWriter{ui}, conn})
			break
		}
	}

	ui.Message("Connection established!")
	return multistep.ActionContinue
}

func (s *stepConnect) Cleanup(state multistep.StateBag) {
	conn, ok := state.GetOk("connection")
	if !ok {
		return
	}

	conn.(sshCommunicator).client.Close()
}

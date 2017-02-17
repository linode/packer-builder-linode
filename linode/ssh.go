package linode

import (
	"github.com/mitchellh/multistep"
	"golang.org/x/crypto/ssh"
)

func commHost(state multistep.StateBag) (string, error) {
	return state.Get("linode_ip").(LinodeIP).Address, nil
}

func sshConfig(state multistep.StateBag) (*ssh.ClientConfig, error) {
	return &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password(state.Get("root_pass").(string)),
		},
	}, nil
}

package linode

import "github.com/mitchellh/multistep"

func commHost(state multistep.StateBag) (string, error) {
	return state.Get("linode_ip").(LinodeIP).Address, nil
}

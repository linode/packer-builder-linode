Linode builder plugin for Packer
================================

A first-draft Packer building for creating Linode images.

## Building

Due to vendoring issues, at the moment this package must be built from within
Packer's source code. So, fetch this repo and Packer:

```sh
$ go get -d github.com/hashicorp/packer
$ go get -d github.com/dradtke/packer-builder-linode
```

Then copy the contents of `linode/` and `test/` to Packer's source tree:

```sh
$ cp -r linode $GOPATH/src/github.com/hashicorp/packer/builder/
$ cp -r test $GOPATH/src/github.com/hashicorp/packer/test
```

Then open up Packer's file `command/plugin.go` and add Linode as a new builder.

```patch
diff --git a/command/plugin.go b/command/plugin.go
index 2d2272640..2d126454c 100644
--- a/command/plugin.go
+++ b/command/plugin.go
@@ -28,6 +28,7 @@ import (
 	hcloudbuilder "github.com/hashicorp/packer/builder/hcloud"
 	hypervisobuilder "github.com/hashicorp/packer/builder/hyperv/iso"
 	hypervvmcxbuilder "github.com/hashicorp/packer/builder/hyperv/vmcx"
+	linodebuilder "github.com/hashicorp/packer/builder/linode"
 	lxcbuilder "github.com/hashicorp/packer/builder/lxc"
 	lxdbuilder "github.com/hashicorp/packer/builder/lxd"
 	ncloudbuilder "github.com/hashicorp/packer/builder/ncloud"
@@ -101,6 +102,7 @@ var Builders = map[string]packer.Builder{
 	"hcloud":              new(hcloudbuilder.Builder),
 	"hyperv-iso":          new(hypervisobuilder.Builder),
 	"hyperv-vmcx":         new(hypervvmcxbuilder.Builder),
+	"linode":              new(linodebuilder.Builder),
 	"lxc":                 new(lxcbuilder.Builder),
 	"lxd":                 new(lxdbuilder.Builder),
 	"ncloud":              new(ncloudbuilder.Builder),
```

Then you can `go install` Packer, and it will have support for the "linode"
plugin.



## Configuration

Check out `test/fixtures/builder-linode/minimal.json` for an example Packer file that uses the
Linode builder. Some notes:

1. You will need a Linode API Token. Naturally it's a bad idea to
   hard-code it, so you will probably want to pull it from an environment
   variable, which is what `minimal.json` does.
1. `ssh_username` is required, generally this value should be `root`.
   All Linode images use `root`, except for `linode/containerlinux` which
   uses `core`.

```sh
$ make dev
$ bin/packer build -var "linode-token=$LINODE_TOKEN" test/fixtures/builder-linode/minimal.json
```

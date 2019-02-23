# Linode builder plugin for Packer

[![GoDoc](https://godoc.org/github.com/dradtke/packer-builder-linode?status.svg)](https://godoc.org/github.com/dradtke/packer-builder-linode)
[![Go Report Card](https://goreportcard.com/badge/github.com/dradtke/packer-builder-linode)](https://goreportcard.com/report/github.com/dradtke/packer-builder-linode)
[![CircleCI](https://circleci.com/gh/dradtke/packer-builder-linode.svg?style=svg)](https://circleci.com/gh/dradtke/packer-builder-linode)
[![GitHub release](https://img.shields.io/github/release/dradtke/packer-builder-linode.svg)](https://github.com/dradtke/packer-builder-linode/releases/)

This is a Packer plug-in for building Linode images.

## Build and Install

```sh
git clone https://github.com/dradtke/packer-builder-linode
cd packer-builder-linode
make install
```

## Configuration

With the plugin installed, check out `test/fixtures/builder-linode/minimal.json` for an example Packer file that uses the
Linode builder.

```
packer build -var "linode-token=$LINODE_TOKEN" test/fixtures/builder-linode/minimal.json
```

Some notes:

1. You will need a Linode APIv4 Personal Access Token.
   Get one here: <https://developers.linode.com/api/v4#section/Personal-Access-Token>
   Naturally it's a bad idea to hard-code it, so you will probably want to pull it from an environment
   variable, which is what `minimal.json` does.
1. `ssh_username` is required, generally this value should be `root`.
   All Linode images use `root`, except for `linode/containerlinux` which
   uses `core`.

## Development

HashiCorp provides excellent guidance on how to build, use, and debug plugins:

* <https://www.packer.io/docs/extending/plugins.html>

Helper tools to interact with the Linode objects exist in the `cmd/` assist in development.
These can be built with `go build` from their path.  They all expect a `LINODE_TOKEN` environment variable to be set.

* `list_images` - Lists all images, both public and private
* `list_events` - Lists the most recent page of events
* `list_kernels` - Lists all kernels
* `delete_image` - Deletes a single image whose *numeric* ID is required (`delete_image 123` to remove image `private/123`)

The [Linode CLI](https://www.linode.com/docs/platform/api/using-the-linode-cli/) can also help during development.
Install it by running `pip install linode-cli`.

### Patching the Packer build tree

These instructions are an alternative to installing `packer-builder-linode` as a plugin.

To patch this plugin into Packer's source code, fetch this repository and Packer:

```sh
go get -d github.com/hashicorp/packer
go get -d github.com/dradtke/packer-builder-linode
```

Then copy the contents of `linode/` and `test/` to Packer's source tree:

```sh
cp -r linode $GOPATH/src/github.com/hashicorp/packer/builder/
cp -r test $GOPATH/src/github.com/hashicorp/packer/test
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

To verify that the Linode patching applied:

```sh
make dev
bin/packer build -var "linode-token=$LINODE_TOKEN" test/fixtures/builder-linode/minimal.json
```

Then you can `go install` Packer, and it will have support for the "linode"
plugin.

## Discussion / Help

Join us at [#linodego](https://gophers.slack.com/messages/CAG93EB2S) on the [gophers slack](https://gophers.slack.com)

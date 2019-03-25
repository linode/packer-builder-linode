# Packer builder plugin for Linode Images

[![GoDoc](https://godoc.org/github.com/linode/packer-builder-linode?status.svg)](https://godoc.org/github.com/linode/packer-builder-linode)
[![Go Report Card](https://goreportcard.com/badge/github.com/linode/packer-builder-linode)](https://goreportcard.com/report/github.com/linode/packer-builder-linode)
[![CircleCI](https://circleci.com/gh/linode/packer-builder-linode.svg?style=svg)](https://circleci.com/gh/linode/packer-builder-linode)
[![GitHub release](https://img.shields.io/github/release/linode/packer-builder-linode.svg)](https://github.com/linode/packer-builder-linode/releases/)

This is a Packer plug-in for building Linode images.

## Work in Progress

This project is currently not supported. Use at your own risk. Contributions welcome!

## Build and Install

Requirements:

* [Packer](https://www.packer.io/intro/getting-started/install.html)
* [Go 1.12+](https://golang.org/doc/install)

Go makes installing the Linode Images builder plugin for Packer easy:

```sh
GOBIN=~/.packer.d/plugins go install github.com/linode/packer-builder-linode
```

To fetch the code and improve the plugin itself:

```sh
git clone https://github.com/linode/packer-builder-linode
cd packer-builder-linode
make install
```

## Configuration

With the builder plugin installed, an Packer file like the example at [`test/fixtures/builder-linode/minimal.json`](https://raw.githubusercontent.com/linode/packer-builder-linode/master/test/fixtures/builder-linode/minimal.json) can create a Linode Image.

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

Helper tools to interact with Linode objects and assist in development can be found in `cmd/`.
These can be built with `go build` within their path.  They all expect a `LINODE_TOKEN` environment variable to be set.

* [`list-images`](cmd/list-images/main.go) - Lists all Linode images, both public and private
* [`list-events`](cmd/list-events/main.go) - Lists the most recent page of events on the Linode account
* [`list-kernels`](cmd/list-kernels/main.go) - Lists all Linode kernels
* [`delete-image`](cmd/delete-image/main.go) - Deletes a single Linode image whose *numeric* ID is required (`delete_image 123` to remove image `private/123`)

The [Linode CLI](https://www.linode.com/docs/platform/api/using-the-linode-cli/) can also help during development.
Install it by running `pip install linode-cli`.

### Patching the Packer build tree

These instructions are an alternative to installing `packer-builder-linode` as a plugin.

To patch this plugin into Packer's source code, fetch this repository and Packer:

```sh
go get -d github.com/hashicorp/packer
go get -d github.com/linode/packer-builder-linode
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

## Contribution Guidelines

Want to improve the Linode Packer Builder? Please start [here](.github/CONTRIBUTING.md).

## Discussion / Help

Join us at [#linodego](https://gophers.slack.com/messages/CAG93EB2S) on the [gophers slack](https://gophers.slack.com)

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

Then copy the contents of `linode/` to Packer's source tree:

```sh
$ cp -r linode $GOPATH/src/github.com/hashicorp/packer/builder/
```

Then open up Packer's file `command/plugin.go` and add Linode as a new builder.
Then you can `go install` Packer, and it will have support for the "linode"
plugin.

## Configuration

Check out `testdata/linode.json` for an example Packer file that uses the
Linode builder. Some notes:

1. You will need a Linode API key. Naturally it's a bad idea to
   hard-code it, so you will probably want to pull it from an environment
   variable, which is what `linode.json` does.
2. A number of values have both a `*_name` and `*_id` variant. If the `*_id`
   version is passed, it expects an integer ID corresponding to that resource
   in Linode; if `*_name` is used, then an API call will be made to determine
   what's available, and pick the one matching the name.
3. `ssh_username` is required, but the only supported value for it is "root".
   Packer's SSH provisioning requires that it be defined, but Linode only
   creates a root user during setup.

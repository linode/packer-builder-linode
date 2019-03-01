PROG := packer-builder-linode
INSTALL_DIR := $${HOME}/.packer.d/plugins

GOOS ?= $(shell go env GOOS)

export GO111MODULE=on

ifeq ($(GOOS),windows)
	BIN_SUFFIX := ".exe"
	INSTALL_DIR := $${APPDATA}/packer.d/plugins
endif

.PHONY: test
test:
	go test -v ./...

.PHONY: build
build:
	go build -o $(PROG)$(BIN_SUFFIX)

.PHONY: check
check:
	go fmt

.PHONY: clean
clean:
	$(RM) $(OUT_DIR)/$(PROG)$(BIN_SUFFIX)

.PHONY: uninstall
uninstall:
	$(RM) $(GOPATH)/bin/$(PROG)$(BIN_SUFFIX)

.PHONY: install
install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(PROG)$(BIN_SUFFIX) $(INSTALL_DIR)

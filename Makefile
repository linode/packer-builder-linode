TEST?=$(shell go list ./...)
PROG := packer-builder-linode
INSTALL_DIR := $${HOME}/.packer.d/plugins

GOOS ?= $(shell go env GOOS)

export GO111MODULE=on

ifeq ($(GOOS),windows)
	BIN_SUFFIX := ".exe"
	INSTALL_DIR := $${APPDATA}/packer.d/plugins
endif

.PHONY: deps fmt fmt-check fmt-docs mode-check test testacc testrace build clean uninstall install

ci: testrace

deps:
	@GO111MODUOE=off go get golang.org/x/tools/cmd/goimports
	@GO111MODUOE=off go get golang.org/x/tools/cmd/stringer
	@GO111MODUOE=off go get -u github.com/mna/pigeon

vet: ## Vet Go code
	@go vet $(VET)  ; if [ $$? -eq 1 ]; then \
		echo "ERROR: Vet found problems in the code."; \
		exit 1; \
	fi

fmt: ## Format Go code
	@gofmt -w -s main.go $(UNFORMATTED_FILES)

fmt-check: ## Check go code formatting
	@echo "==> Checking that code complies with gofmt requirements..."
	@if [ ! -z "$(UNFORMATTED_FILES)" ]; then \
		echo "gofmt needs to be run on the following files:"; \
		echo "$(UNFORMATTED_FILES)" | xargs -n1; \
		echo "You can use the command: \`make fmt\` to reformat code."; \
		exit 1; \
		else \
		echo "Check passed."; \
		fi

mode-check: ## Check that only certain files are executable
	@echo "==> Checking that only certain files are executable..."
	@if [ ! -z "$(EXECUTABLE_FILES)" ]; then \
		echo "These files should not be executable or they must be white listed in the Makefile:"; \
		echo "$(EXECUTABLE_FILES)" | xargs -n1; \
		exit 1; \
		else \
		echo "Check passed."; \
		fi
fmt-docs:
	@find ./website/source/docs -name "*.md" -exec pandoc --wrap auto --columns 79 --atx-headers -s -f "markdown_github+yaml_metadata_block" -t "markdown_github+yaml_metadata_block" {} -o {} \;

test: fmt-check mode-check vet ## Run unit tests
	@go test $(TEST) $(TESTARGS) -timeout=3m

# testacc runs acceptance tests
testacc: deps ## Run acceptance tests
	@echo "WARN: Acceptance tests will take a long time to run and may cost money. Ctrl-C if you want to cancel."
	PACKER_ACC=1 go test -v $(TEST) $(TESTARGS) -timeout=45m

testrace: fmt-check mode-check vet ## Test with race detection enabled
	@GO111MODULE=off go test -race $(TEST) $(TESTARGS) -timeout=3m -p=8

build:
	go build -o $(PROG)$(BIN_SUFFIX)

check:
	go fmt

clean:
	$(RM) $(OUT_DIR)/$(PROG)$(BIN_SUFFIX)

uninstall:
	$(RM) $(GOPATH)/bin/$(PROG)$(BIN_SUFFIX)

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(PROG)$(BIN_SUFFIX) $(INSTALL_DIR)

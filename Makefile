.PHONY: build test lint install install-cli uninstall

INSTALLER ?= bin/sdlc-install
PROJECT_INITIALIZER ?= bin/sdlc-project-init
INSTALL_FLAGS ?=

build:
	go build -o $(INSTALLER) ./cmd/sdlc-install
	go build -o $(PROJECT_INITIALIZER) ./cmd/sdlc-project-init

test: lint
	go test ./...

lint:
	go vet ./...
	golangci-lint run ./...
	shellcheck hooks/agent-command-guard.sh
	shfmt -i 4 -d hooks/agent-command-guard.sh

install: install-cli
	@$(INSTALLER) $(INSTALL_FLAGS)

install-cli: build
	mkdir -p "$(HOME)/.local/bin"
	ln -sfn "$(CURDIR)/$(INSTALLER)" "$(HOME)/.local/bin/sdlc-install"
	ln -sfn "$(CURDIR)/$(PROJECT_INITIALIZER)" "$(HOME)/.local/bin/sdlc-project-init"

uninstall:
	trash "$(HOME)/.local/bin/sdlc-install"
	trash "$(HOME)/.local/bin/sdlc-project-init"

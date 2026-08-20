.PHONY: build test lint install install-cli uninstall

INSTALLER ?= bin/sdlc-install
INSTALL_FLAGS ?=

build:
	go build -o bin/sdlc-install ./cmd/sdlc-install

test: lint
	go test ./...

lint:
	go vet ./...
	golangci-lint run ./...
	shellcheck hooks/agent-command-guard.sh
	shfmt -i 4 -d hooks/agent-command-guard.sh

install: build
	@$(INSTALLER) $(INSTALL_FLAGS)

install-cli: build
	mkdir -p "$(HOME)/.local/bin"
	ln -s "$(CURDIR)/bin/sdlc-install" "$(HOME)/.local/bin/sdlc-install"

uninstall:
	trash "$(HOME)/.local/bin/sdlc-install"

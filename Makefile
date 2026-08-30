.PHONY: build test lint install install-cli uninstall

INSTALLER ?= bin/sdlc-install
PROJECT_INITIALIZER ?= bin/sdlc-project-init
INSTALL_FLAGS ?=
SDLC_RELEASE ?= $(shell git for-each-ref --count=1 --sort=-version:refname --format='%(refname:short)' --points-at=HEAD 'refs/tags/v*')
PROJECT_INITIALIZER_BUILD_FLAGS :=
ifneq ($(strip $(SDLC_RELEASE)),)
PROJECT_INITIALIZER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
endif

build:
	go build -o $(INSTALLER) ./cmd/sdlc-install
	go build $(PROJECT_INITIALIZER_BUILD_FLAGS) -o $(PROJECT_INITIALIZER) ./cmd/sdlc-project-init

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

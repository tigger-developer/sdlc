.PHONY: build test lint install install-cli uninstall

INSTALLER ?= bin/sdlc-install
PROJECT_INITIALIZER ?= bin/sdlc-init
PREVIEWER ?= bin/sdlc-preview
LEGACY_AC_MERGER ?= bin/sdlc-merge-legacy-acs
INSTALL_FLAGS ?=
SDLC_RELEASE ?= $(shell git for-each-ref --merged=HEAD --count=1 --sort=-version:refname --format='%(refname:short)' 'refs/tags/v*')
INSTALLER_BUILD_FLAGS :=
PROJECT_INITIALIZER_BUILD_FLAGS :=
PREVIEWER_BUILD_FLAGS :=
LEGACY_AC_MERGER_BUILD_FLAGS :=
ifneq ($(strip $(SDLC_RELEASE)),)
INSTALLER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
PROJECT_INITIALIZER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
PREVIEWER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
LEGACY_AC_MERGER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
endif

build:
	go build $(INSTALLER_BUILD_FLAGS) -o $(INSTALLER) ./cmd/sdlc-install
	go build $(PROJECT_INITIALIZER_BUILD_FLAGS) -o $(PROJECT_INITIALIZER) ./cmd/sdlc-init
	go build $(PREVIEWER_BUILD_FLAGS) -o $(PREVIEWER) ./cmd/sdlc-preview
	go build $(LEGACY_AC_MERGER_BUILD_FLAGS) -o $(LEGACY_AC_MERGER) ./cmd/sdlc-merge-legacy-acs

test: lint
	go test ./...

lint:
	go vet ./...
	golangci-lint run ./...
	shellcheck hooks/agent-command-guard.sh src/libexec/load-sdlc-env.sh
	shfmt -i 4 -d hooks/agent-command-guard.sh src/libexec/load-sdlc-env.sh

install: build
	@$(INSTALLER) $(INSTALL_FLAGS)

install-cli: install

uninstall:
	unlink "$(HOME)/.local/bin/sdlc-init"
	unlink "$(HOME)/.local/bin/sdlc-preview"
	unlink "$(HOME)/.local/bin/sdlc-merge-legacy-acs"

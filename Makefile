.PHONY: build test lint install install-cli uninstall

INSTALLER ?= bin/sdlc-install
PROJECT_INITIALIZER ?= bin/sdlc-project-init
PROJECT_UPDATER ?= bin/sdlc-project-update
AUDIT_RUNNER ?= bin/sdlc-audit
INSTALL_FLAGS ?=
SDLC_RELEASE ?= $(shell git for-each-ref --count=1 --sort=-version:refname --format='%(refname:short)' --points-at=HEAD 'refs/tags/v*')
PROJECT_INITIALIZER_BUILD_FLAGS :=
AUDIT_RUNNER_BUILD_FLAGS :=
ifneq ($(strip $(SDLC_RELEASE)),)
PROJECT_INITIALIZER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
AUDIT_RUNNER_BUILD_FLAGS += -ldflags "-X main.buildVersion=$(SDLC_RELEASE)"
endif

build:
	go build -o $(INSTALLER) ./cmd/sdlc-install
	go build $(PROJECT_INITIALIZER_BUILD_FLAGS) -o $(PROJECT_INITIALIZER) ./cmd/sdlc-project-init
	go build $(PROJECT_INITIALIZER_BUILD_FLAGS) -o $(PROJECT_UPDATER) ./cmd/sdlc-project-init
	go build $(AUDIT_RUNNER_BUILD_FLAGS) -o $(AUDIT_RUNNER) ./cmd/sdlc-audit

test: lint
	go test ./...

lint:
	go vet ./...
	golangci-lint run ./...
	shellcheck hooks/agent-command-guard.sh src/libexec/load-sdlc-env.sh
	shfmt -i 4 -d hooks/agent-command-guard.sh src/libexec/load-sdlc-env.sh

install: install-cli
	@$(INSTALLER) $(INSTALL_FLAGS)

install-cli: build
	mkdir -p "$(HOME)/.local/bin"
	ln -sfn "$(CURDIR)/$(INSTALLER)" "$(HOME)/.local/bin/sdlc-install"
	ln -sfn "$(CURDIR)/$(PROJECT_INITIALIZER)" "$(HOME)/.local/bin/sdlc-project-init"
	ln -sfn "$(CURDIR)/$(PROJECT_UPDATER)" "$(HOME)/.local/bin/sdlc-project-update"
	ln -sfn "$(CURDIR)/$(AUDIT_RUNNER)" "$(HOME)/.local/bin/sdlc-audit"

uninstall:
	trash "$(HOME)/.local/bin/sdlc-install"
	trash "$(HOME)/.local/bin/sdlc-project-init"
	trash "$(HOME)/.local/bin/sdlc-project-update"
	trash "$(HOME)/.local/bin/sdlc-audit"

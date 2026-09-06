.PHONY: build test lint install install-cli uninstall

INSTALLER ?= bin/sdlc-install
PROJECT_INITIALIZER ?= bin/sdlc-project-init
PREVIEWER ?= bin/sdlc-preview
LEGACY_AC_MERGER ?= bin/sdlc-merge-legacy-acs
INSTALL_FLAGS ?=
SDLC_RELEASE ?= $(shell git for-each-ref --count=1 --sort=-version:refname --format='%(refname:short)' --points-at=HEAD 'refs/tags/v*')
PROJECT_INITIALIZER_BUILD_FLAGS :=
PREVIEWER_BUILD_FLAGS :=
LEGACY_AC_MERGER_BUILD_FLAGS :=
ifneq ($(strip $(SDLC_RELEASE)),)
PROJECT_INITIALIZER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
PREVIEWER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
LEGACY_AC_MERGER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
endif

build:
	go build -o $(INSTALLER) ./cmd/sdlc-install
	go build $(PROJECT_INITIALIZER_BUILD_FLAGS) -o $(PROJECT_INITIALIZER) ./cmd/sdlc-project-init
	go build $(PREVIEWER_BUILD_FLAGS) -o $(PREVIEWER) ./cmd/sdlc-preview
	go build $(LEGACY_AC_MERGER_BUILD_FLAGS) -o $(LEGACY_AC_MERGER) ./cmd/sdlc-merge-legacy-acs

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
	@path="$(HOME)/.local/bin/sdlc-project-update"; if [ -e "$$path" ] || [ -L "$$path" ]; then backup="$$path.sdlc-v2-retired"; suffix=2; while [ -e "$$backup" ] || [ -L "$$backup" ]; do backup="$$path.sdlc-v2-retired-$$suffix"; suffix=$$((suffix + 1)); done; mv "$$path" "$$backup"; echo "Retired $$path -> $$backup"; fi
	@path="$(HOME)/.local/bin/sdlc-audit"; if [ -e "$$path" ] || [ -L "$$path" ]; then backup="$$path.sdlc-v2-retired"; suffix=2; while [ -e "$$backup" ] || [ -L "$$backup" ]; do backup="$$path.sdlc-v2-retired-$$suffix"; suffix=$$((suffix + 1)); done; mv "$$path" "$$backup"; echo "Retired $$path -> $$backup"; fi
	@target="$(CURDIR)/$(INSTALLER)"; path="$(HOME)/.local/bin/sdlc-install"; if [ ! -L "$$path" ] || [ "$$(readlink "$$path")" != "$$target" ]; then if [ -e "$$path" ] || [ -L "$$path" ]; then backup="$$path.sdlc-retired"; suffix=2; while [ -e "$$backup" ] || [ -L "$$backup" ]; do backup="$$path.sdlc-retired-$$suffix"; suffix=$$((suffix + 1)); done; mv "$$path" "$$backup"; echo "Backed up $$path -> $$backup"; fi; ln -s "$$target" "$$path"; echo "Installed $$path"; fi
	@target="$(CURDIR)/$(PROJECT_INITIALIZER)"; path="$(HOME)/.local/bin/sdlc-project-init"; if [ ! -L "$$path" ] || [ "$$(readlink "$$path")" != "$$target" ]; then if [ -e "$$path" ] || [ -L "$$path" ]; then backup="$$path.sdlc-retired"; suffix=2; while [ -e "$$backup" ] || [ -L "$$backup" ]; do backup="$$path.sdlc-retired-$$suffix"; suffix=$$((suffix + 1)); done; mv "$$path" "$$backup"; echo "Backed up $$path -> $$backup"; fi; ln -s "$$target" "$$path"; echo "Installed $$path"; fi
	@target="$(CURDIR)/$(PREVIEWER)"; path="$(HOME)/.local/bin/sdlc-preview"; if [ ! -L "$$path" ] || [ "$$(readlink "$$path")" != "$$target" ]; then if [ -e "$$path" ] || [ -L "$$path" ]; then backup="$$path.sdlc-retired"; suffix=2; while [ -e "$$backup" ] || [ -L "$$backup" ]; do backup="$$path.sdlc-retired-$$suffix"; suffix=$$((suffix + 1)); done; mv "$$path" "$$backup"; echo "Backed up $$path -> $$backup"; fi; ln -s "$$target" "$$path"; echo "Installed $$path"; fi
	@target="$(CURDIR)/$(LEGACY_AC_MERGER)"; path="$(HOME)/.local/bin/sdlc-merge-legacy-acs"; if [ ! -L "$$path" ] || [ "$$(readlink "$$path")" != "$$target" ]; then if [ -e "$$path" ] || [ -L "$$path" ]; then backup="$$path.sdlc-retired"; suffix=2; while [ -e "$$backup" ] || [ -L "$$backup" ]; do backup="$$path.sdlc-retired-$$suffix"; suffix=$$((suffix + 1)); done; mv "$$path" "$$backup"; echo "Backed up $$path -> $$backup"; fi; ln -s "$$target" "$$path"; echo "Installed $$path"; fi

uninstall:
	unlink "$(HOME)/.local/bin/sdlc-install"
	unlink "$(HOME)/.local/bin/sdlc-project-init"
	unlink "$(HOME)/.local/bin/sdlc-preview"
	unlink "$(HOME)/.local/bin/sdlc-merge-legacy-acs"

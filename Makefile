.PHONY: build test lint install install-preview install-cli uninstall sync

INSTALLER ?= bin/sdlc-install
PROJECT_INITIALIZER ?= bin/sdlc-init
LEGACY_AC_MERGER ?= bin/sdlc-merge-legacy-acs
HARNESS_RUNNER ?= bin/sdlc-harness
INSTALL_FLAGS ?=
COMMIT_MESSAGE ?= chore: sync
export COMMIT_MESSAGE
SDLC_RELEASE ?= $(shell git for-each-ref --merged=HEAD --count=1 --sort=-version:refname --format='%(refname:short)' 'refs/tags/v*')
INSTALLER_BUILD_FLAGS :=
PROJECT_INITIALIZER_BUILD_FLAGS :=
LEGACY_AC_MERGER_BUILD_FLAGS :=
HARNESS_RUNNER_BUILD_FLAGS :=
ifneq ($(strip $(SDLC_RELEASE)),)
INSTALLER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
PROJECT_INITIALIZER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
LEGACY_AC_MERGER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
HARNESS_RUNNER_BUILD_FLAGS += -ldflags "-X main.buildRelease=$(SDLC_RELEASE)"
endif

build:
	go build $(INSTALLER_BUILD_FLAGS) -o $(INSTALLER) ./cmd/sdlc-install
	go build $(PROJECT_INITIALIZER_BUILD_FLAGS) -o $(PROJECT_INITIALIZER) ./cmd/sdlc-init
	go build $(LEGACY_AC_MERGER_BUILD_FLAGS) -o $(LEGACY_AC_MERGER) ./cmd/sdlc-merge-legacy-acs
	go build $(HARNESS_RUNNER_BUILD_FLAGS) -o $(HARNESS_RUNNER) ./cmd/sdlc-harness

test: lint
	go test ./...

lint:
	go vet ./...
	golangci-lint run ./...
	shellcheck hooks/agent-command-guard.sh src/libexec/load-sdlc-env.sh
	shfmt -i 4 -d hooks/agent-command-guard.sh src/libexec/load-sdlc-env.sh

install: build
	@$(INSTALLER) $(INSTALL_FLAGS)
	@$(MAKE) install-preview

install-preview:
	git submodule update --init -- tools/HTML-Preview
	@if command -v htmlpreview >/dev/null 2>&1; then \
		printf '%s\n' 'HTML-Preview: htmlpreview is already installed.'; \
	else \
		$(MAKE) -C tools/HTML-Preview install; \
	fi

install-cli: install

sync:
	git submodule update --init --remote -- tools/HTML-Preview
	$(MAKE) -C tools/HTML-Preview install
	git add -A
	@git diff --cached --quiet; result=$$?; \
	case $$result in \
		0) ;; \
		1) git commit -m "$$COMMIT_MESSAGE" ;; \
		*) exit $$result ;; \
	esac
	git pull
	git push

uninstall:
	unlink "$(HOME)/.local/bin/sdlc-init"
	unlink "$(HOME)/.local/bin/sdlc-merge-legacy-acs"

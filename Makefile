.PHONY: build test lint install uninstall

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
	mkdir -p "$(HOME)/.local/bin"
	ln -s "$(CURDIR)/bin/sdlc-install" "$(HOME)/.local/bin/sdlc-install"

uninstall:
	trash "$(HOME)/.local/bin/sdlc-install"

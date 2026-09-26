BIN := $(HOME)/.local/bin/claude-statusline
SAMPLE := $(HOME)/.claude/statusline-sample.json

.PHONY: build install preview test

build:
	CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o claude-statusline .

install: build
	install -Dm755 claude-statusline $(BIN)

preview: build
	./claude-statusline < $(SAMPLE); echo

test:
	go test ./...

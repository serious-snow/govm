GOOS:=$(shell go env GOOS)


.PHONY: build
build:
	go build -ldflags "-s -w" -trimpath -o ./ ./cmds/govm/

.PHONY: lint
lint:
	golangci-lint run -v

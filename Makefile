.PHONY: all
# generate all
build:
	go build -ldflags "-s -w" -trimpath -o govm

.PHONY: lint
lint:
	golangci-lint run -v

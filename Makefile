.PHONY: build
build:
	go build -ldflags "-s -w" -trimpath -o ./ ./cmds/govm/

.PHONY: lint
# 代码检查
lint:
	mise x golangci-lint@2.6.2 -- golangci-lint run -v --timeout=10m  --allow-parallel-runners

.PHONY: fmt
# 代码格式化
fmt:
	mise x golangci-lint@2.6.2 -- golangci-lint fmt

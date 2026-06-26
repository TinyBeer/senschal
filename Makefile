APP_NAME = seneschal
BUILD_DIR = build

# Go 基础环境配置
GO := go
GO_MOD := $(GO) mod
GO_INSTALL := $(GO) install
GO_GENERATE := $(GO) generate

# 工具定义：path@version
TOOL_LIST := \
	golang.org/x/tools/cmd/stringer@v0.35.0\
	mvdan.cc/gofumpt@v0.10.0

.PHONY: gen build clean format tools test

# 安装所有依赖插件
tools:
# 	@echo "=== 开始安装依赖工具 ==="
	$(foreach tool,$(TOOL_LIST),$(GO_INSTALL) $(tool);)
# 	@echo "=== 依赖工具安装完成 ==="

gen: 
	@go generate ./...

format: 
	@gofumpt -w --extra .

# 编译项目
build: format gen
	@go build -o $(BUILD_DIR)/$(APP_NAME) .

# 测试
test: format gen
	@go test -cover --count=1 ./...

clean:
	@rm -rf $(BUILD_DIR)
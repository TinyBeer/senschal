APP_NAME = seneschal
BUILD_DIR = data/build

# Go 基础环境配置
GO := go
GO_MOD := $(GO) mod
GO_INSTALL := $(GO) install
GO_GENERATE := $(GO) generate

# 工具定义：path@version
TOOL_LIST := \
	golang.org/x/tools/cmd/stringer@v0.35.0

.PHONY: gen build build-all clean

# 安装所有依赖插件
.PHONY: tools
tools:
# 	@echo "=== 开始安装依赖工具 ==="
	$(foreach tool,$(TOOL_LIST),$(GO_INSTALL) $(tool);)
# 	@echo "=== 依赖工具安装完成 ==="

.PHONY: gen
gen: tools
	@go generate ./...

# 编译项目
.PHONY: build
build: gen
	@go build -o $(BUILD_DIR)/$(APP_NAME) .

clean:
	@rm -rf $(BUILD_DIR)
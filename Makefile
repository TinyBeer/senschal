APP_NAME  = seneschal
MAIN_FILE = .
OUT_DIR   = out

# Go 环境变量
GO           := go
GO_INSTALL   := $(GO) install
GO_MOD       := $(GO) mod
GO_BUILD     := $(GO) build
GO_TEST      := $(GO) test
GO_VET       := $(GO) vet
GO_FMT       := gofumpt
GO_FMT_FLAGS := -w -extra
GOLINT       := golangci-lint

# 编译参数
CGO_ENABLED  := 0
BUILD_FLAGS  := -ldflags "-s -w"
# LDFLAGS_VERSION := -X $(MODULE_PATH)/internal/version.Version=$(shell git describe --tags --always)

# 工具定义：path@version
TOOL_LIST := \
	golang.org/x/tools/cmd/stringer@v0.35.0 \
	mvdan.cc/gofumpt@v0.10.0 \
	github.com/golangci/golangci-lint@v1.64.8

# 交叉编译目标平台
PLATFORMS := \
	linux/amd64 \
	linux/arm64 \
	darwin/amd64 \
	darwin/arm64 \
	windows/amd64

# 颜色输出
RED    := \033[31m
GREEN  := \033[32m
YELLOW := \033[33m
END    := \033[0m

# 默认目标：编译当前系统二进制
.PHONY: all
all: gen fmt vet lint test build

# 1. 依赖管理
.PHONY: tidy
tidy:
	@echo -e "$(GREEN)=== 整理go mod依赖 ===$(END)"
	$(GO_MOD) tidy
	$(GO_MOD) verify

.PHONY: download
download:
	@echo -e "$(GREEN)=== 下载全部依赖 ===$(END)"
	$(GO_MOD) download

# 2. 代码生成 & 代码格式化 & 静态检查
.PHONY: gen generate
gen: generate # 别名
generate: tidy
	@echo -e "$(GREEN)=== go generate 自动生成代码 ===$(END)"
	$(GO) generate ./...
	# 生成后自动格式化生成的代码
	$(GO_FMT) $(GO_FMT_FLAGS) .

.PHONY: fmt
fmt:
	@echo -e "$(GREEN)=== 格式化代码 ===$(END)"
	$(GO_FMT) $(GO_FMT_FLAGS) .

.PHONY: vet
vet:
	@echo -e "$(GREEN)=== go vet 静态检查 ===$(END)"
	$(GO_VET) -all ./...

.PHONY: lint
lint:
	@echo -e "$(GREEN)=== golangci-lint 代码规范检查 ===$(END)"
	$(GOLINT) run ./...

# 3. 单元测试
.PHONY: mk_out
mk_out:
	@mkdir -p $(OUT_DIR)

.PHONY: test
test: mk_out
	@echo -e "$(GREEN)=== 执行单元测试 ===$(END)"
	$(GO_TEST) -v -race -coverprofile=$(OUT_DIR)/coverage.out ./...

.PHONY: cover
cover: test
	@echo -e "$(GREEN)=== 生成覆盖率报告 ===$(END)"
	$(GO) tool cover -html=$(OUT_DIR)/coverage.out

# 4. 编译
.PHONY: build
build: mk_out
	@echo -e "$(GREEN)=== 编译当前平台二进制 ===$(END)"
	mkdir -p $(OUT_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO_BUILD) $(BUILD_FLAGS) -o $(OUT_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo -e "$(GREEN)输出: $(OUT_DIR)/$(APP_NAME)$(END)"

.PHONY: cross
cross: mk_out
	@echo -e "$(GREEN)=== 全平台交叉编译 ===$(END)"
	mkdir -p $(OUT_DIR)
	$(foreach platform,$(PLATFORMS),\
		GOOS=$(word 1,$(subst /, ,$(platform))) \
		GOARCH=$(word 2,$(subst /, ,$(platform))) \
		CGO_ENABLED=$(CGO_ENABLED) \
		$(GO_BUILD) $(BUILD_FLAGS) -o $(OUT_DIR)/$(APP_NAME)-$(subst /,-,$(platform))$(if $(findstring windows,$(platform)),.exe) $(MAIN_FILE);)
	@echo -e "$(GREEN)输出目录: $(OUT_DIR)$(END)"

# 5. 清理产物
.PHONY: clean
clean:
	@echo -e "$(YELLOW)=== 清理编译产物 ===$(END)"
	rm -rf $(OUT_DIR)

# 6. 安装工具链
.PHONY: install-tools
install-tools:
	@echo -e "$(GREEN)=== 安装开发工具 ===$(END)"
	$(foreach tool,$(TOOL_LIST),$(GO_INSTALL) $(tool);)

# 帮助文档
.PHONY: help
help:
	@echo -e "$(YELLOW)可用命令列表:$(END)"
	@echo -e "  make tidy         整理&校验依赖"
	@echo -e "  make download      下载全部依赖"
	@echo -e "  make fmt          格式化代码"
	@echo -e "  make vet          go vet 静态代码检查"
	@echo -e "  make lint         golangci-lint 规范校验"
	@echo -e "  make test         执行单元测试+覆盖率"
	@echo -e "  make cover        打开html覆盖率报告"
	@echo -e "  make build        编译当前系统二进制"
	@echo -e "  make cross        全平台交叉编译"
	@echo -e "  make clean        清理编译产物"
	@echo -e "  make install-tools 安装开发工具"
	@echo -e "  make gen          执行 go generate"
	@echo -e "  make all          完整检查+编译流程"

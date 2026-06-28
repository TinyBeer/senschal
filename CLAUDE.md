# CLAUDE.md

此文件为 Claude Code (claude.ai/code) 在操作此仓库时提供指引。

## 项目概述

Seneschal 是一个 Go CLI 工具，用于在云服务器/Linux 服务器上自动化环境部署。它管理 Docker 环境、protobuf 服务脚手架、基于 SSH 的文件传输、图片转 ASCII 艺术字，以及一个运动计时器 TUI。

## 编码流程

### 确认背景

首先阅读 README、配置文件和目录结构，确认项目目标、技术栈和关键目录；不要凭经验猜。

### 明确需求

不要假装已经理解，也不要把不确定性藏起来。该说清楚的地方先说清楚。

- 明确写出你的假设；不确定就问。
- 如果需求存在多种理解，把几种理解列出来，不要悄悄选一个。
- 如果有更简单的做法，要主动说明。
- 该反对时要反对；发现需求不清楚，就停下来指出哪里不清楚。

### 代码设计简单优先

1. 对于新增需求，只写解决当前问题所需的最少代码，不做猜测式扩展。
   - 不添加用户没有要求的功能。
   - 不为只用一次的代码设计抽象层。
   - 不加入没有被要求的灵活性、配置项或扩展点。
   - 不为不可能出现的场景堆错误处理。
   - 如果 200 行能改成 50 行，就回头简化。

   自检问题：资深工程师看到这段实现，会不会觉得过度设计？如果会，就删减。

2. 对于代码修改，外科手术式修改，只碰必须修改的地方。只清理自己造成的问题。
   - 不顺手"优化"旁边的代码、注释或格式。
   - 不重构没有坏的部分。
   - 匹配项目现有风格，即使你个人会用另一种写法。
   - 发现无关的废代码可以提一句，不要直接删除。

   如果你的改动制造了无用内容：
   - 删除由本次改动造成的无用 import、变量、函数。
   - 不删除本来就存在的死代码，除非任务明确要求。

   检查标准：每一行改动都应该能追溯到用户这次请求。

### 具体执行

开发流程遵循 TDD，使用 `github.com/stretchr/testify` 测试框架：
1. 红：先写单元测试（预期失败）
2. 绿：写刚好通过测试的最少业务代码
3. 重构：测试全绿后重构，重构不改动测试

多步骤任务先写简短计划：

```text
1. [步骤] -> 验证：[检查方式]
2. [步骤] -> 验证：[检查方式]
3. [步骤] -> 验证：[检查方式]
```

示例：

- "加校验" -> "先写非法输入测试，再让测试通过"
- "修 bug" -> "先写复现 bug 的测试，再修到测试通过"
- "重构 X" -> "重构前后都确认测试通过"

## 编码规范

### Cobra

1. rootCmd 统一初始化全局flag（verbose、config、output）
2. 子命令拆分独立文件，单一命令一个文件
3. flag 区分持久flag（全局）、本地flag（单命令）
4. 添加命令示例 `cmd.Flags().Example = ""`
5. 命令参数校验在 PreRunE，业务逻辑 RunE
6. 所有错误在 RunE 返回，统一 cobra 错误打印

### 输出规范

1. 区分 stdout（正常输出json/table）、stderr（日志错误）
2. 支持 --output json/table/yaml 多格式输出
3. 静默模式 --quiet 关闭非必要打印
4. 进度条、交互式输入封装工具包，不重复实现

### 文件与IO

1. 文件操作统一使用 os + afero 抽象，方便单元测试
2. 文件读写携带上下文，处理权限、不存在异常
3. 大文件流式读取，禁止一次性加载全量到内存
4. 路径处理使用 filepath 包，不硬编码 / 或 \

## 构建与开发

```bash
# 编译
go build -o seneschal .

# 运行
go run main.go [command]

# 生成代码（stringer 枚举）
go generate ./...

# 运行测试
go test ./...

# 运行单个包测试
go test ./ui/component/...

# 整理依赖
go mod tidy
```

## CLI 命令

| 命令                                            | 说明                                                                                       |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------ |
| `seneschal`                                     | 根命令 — 列出已注册命令                                                                    |
| `seneschal joy`                                 | Joynova 项目工具：列出项目、注册 protobuf 接口（`joy inter`）、从模板生成文件（`joy tpl`） |
| `seneschal img text <file>`                     | 图片转 ASCII 艺术字（参数：`-W` 宽度，`-H` 高度，`-V` 反转，`-C` ANSI 颜色）               |
| `seneschal img edge <file.gif>`                 | 对 GIF 应用 Sobel 边缘检测                                                                 |
| `seneschal env`                                 | 列出环境配置                                                                               |
| `seneschal cp <src> <dst>`                      | 在本地和远程之间复制文件/目录，支持 `alias:path` 语法                                     |
| `seneschal agent list`                          | 列出 SSH 代理（服务器）配置                                                                |
| `seneschal agent up <aliases> <local> <remote>` | 上传文件/目录到多个代理，别名逗号分隔                                                      |
| `seneschal agent down <alias> <remote> <local>` | 从代理下载文件/目录到本地                                                                  |
| `seneschal agent check <aliases> <env>`         | 检查远程代理的环境状态                                                                     |
| `seneschal agent deploy <aliases> <env>`        | 部署环境（Docker）到远程代理                                                               |
| `seneschal todo [add/done/del]`                 | Todo 列表管理器，JSON 文件持久化                                                           |
| `seneschal workout [-l] [name]`                 | 使用 Bubble Tea TUI 的运动计时器，CSV 配置运动计划                                         |

## 项目结构

```text
seneschal/
├── main.go                # 入口，调用 cmd.Execut()
├── cmd/                   # Cobra 命令定义
│   ├── root.go            # 根命令
│   ├── joy.go             # Joynova 项目与模板工具
│   ├── img.go             # 图片处理（ASCII 艺术字、边缘检测）
│   ├── env.go             # 环境列表
│   ├── agent.go           # SSH 代理管理、检查、部署
│   ├── cp.go              # 本地与远程之间复制文件/目录
│   ├── todo.go            # Todo CRUD 命令
│   ├── workout.go         # 运动命令
│   └── test.go            # 开发测试沙盒
├── config/                # 配置类型与加载器
│   ├── const.go           # 目录路径常量（data/conf/、data/tpl/ 等）
│   ├── env.go             # 环境配置、Docker 配置（版本、镜像、用户组）
│   ├── project.go         # 项目配置（proto 目录、服务目录、lobby 注册）
│   ├── ssh.go             # SSH 配置（密码/密钥认证、主机、端口）
│   ├── workout.go         # 从 CSV 读取的运动配置（时长/计数/休息项）
│   └── workouttype_string.go  # 自动生成的 stringer
├── internal/
│   ├── fsutil/             # 文件系统抽象层（本地/远程统一接口）
│   │   ├── core.go         # FileSystem 接口、PathRef、localFS、remoteFS、Transfer（CopyFile/CopyDir/Upload/Download）
│   │   ├── core_test.go    # 本地文件操作测试
│   │   ├── localfs.go      # localClient 本地文件操作底层封装
│   │   ├── sshfs.go        # SSH 客户端封装与 RemoteClient 接口
│   │   ├── sshfs_test.go   # remoteFS Mock 测试
│   │   └── sshclient_test.go # sshClient 真实 SSH 连接单元测试
│   ├── runner/             # 执行逻辑，与命令解耦
│   │   ├── exec.go         # 通过 os/exec 执行本地命令
│   │   ├── ssh.go          # SSH 客户端、会话管理、断点续传
│   │   ├── docker.go       # Docker 镜像拉取、保存与加载编排
│   │   └── env_mgr/        # 环境管理器接口与实现
│   │       ├── common.go   # IEnvMgr 接口（Check、Deploy）
│   │       └── docker.go   # 在 SSH 代理上检查/部署 Docker
│   └── command/            # 命令业务实现
│       ├── file/           # 文件子操作（模板、proto 代码注入）
│       │   ├── const.go    # 文件扩展名、ReplaceProbe 枚举
│       │   ├── list.go     # 遍历目录、按扩展名列出文件
│       │   ├── read.go     # 检查文件是否包含指定字符串
│       │   ├── write.go    # 在探针标记处向文件插入代码
│       │   ├── proto.go    # 解析 .proto 文件、生成 message 与 RPC
│       │   └── tpl.go      # 使用 TOML 设置的 Go 模板执行
│       ├── img/            # 图片处理
│       │   ├── text.go     # 图片转 ASCII（可配置字符集、颜色）
│       │   └── edge.go     # GIF 帧的 Sobel 边缘检测
│       └── todo/           # Todo 领域对象与 JSON 文件仓库
│           ├── model.go    # Todo 结构体与状态枚举
│           ├── domain.go   # TodoRepository 接口、单例获取
│           └── repository.go # TodoFileRepo（JSON CRUD）
├── pkg/util/               # 工具函数
│   ├── table.go            # 支持 markdown 的 CLI 表格渲染
│   └── file.go             # 文件保存工具
└── ui/
     ├── component/         # 终端 UI 组件库
     │   ├── container.go  # Box（水平/垂直布局）、InlineText、Rectangle 容器
     │   ├── style.go       # Lipgloss 样式预设
     │   ├── string_util.go # 支持中文的滑动窗口字符串显示
     │   └── string_util_test.go
     └── terminal/
         └── workout.go     # 基于 Bubble Tea 的运动计时器 TUI

```

## 关键设计模式

### 配置加载

所有配置类型（SSH、Env、Project）遵循相同模式：TOML 文件存放在 `data/conf/` 子目录中，通过 `viper` 加载，使用 `file.ListFileNameWithExt()` + `read*FromToml()` 函数，返回以别名为键的 map。

### 统一的本地/远程路径

`internal/fsutil.PathRef` 通过 `alias:path` 地址方案（例如 `myserver:/tmp/file.txt`）支持无缝的本地和 SSH 远程操作。通过 `fsutil.Parse()` 解析路径，`fsutil.GetFS()` 获取对应的文件系统接口。参见 `internal/fsutil/core.go`、`internal/fsutil/sshfs.go`。

### 通过探针标记进行代码注入

`internal/command/file` 包支持将生成的 protobuf message/RPC 定义插入到现有的 `.proto` 文件中。文件必须包含 `//rpc place`、`//message place` 或 `//func place` 探针标记。参见 `internal/command/file/write.go` 和 `cmd/joy.go`。

### 环境管理器接口

`internal/runner/env_mgr/common.go` 中的 `IEnvMgr` 定义了 `Check()` 和 `Deploy()` 方法。目前仅实现了 `EnvMgrDocker`，它安装 Docker（通过网络或 .deb）、配置用户组并加载所需的 Docker 镜像。

### 模板引擎

模板使用 Go 的 `text/template`，设置来自 `setting.toml` 文件。模板位于 `data/tpl/<name>/template/` 并生成到 `data/_gen/<name>/`。参见 `internal/command/file/tpl.go` 和 `cmd/joy.go` 中的 `joy tpl exec`。

## 数据约定

- **配置格式**：SSH/project/env 配置使用 TOML；运动计划使用 CSV；Todo 持久化使用 JSON
- **配置目录**：`data/conf/<type>/`，其中 type 为 `env`、`project`、`ssh`、`workout`
- **远程路径格式**：`ssh_alias:/path/to/file` — `fsutil.Parse()` 解析此格式

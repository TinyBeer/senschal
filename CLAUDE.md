# CLAUDE.md

## 项目概述

Seneschal — Go CLI 工具，用于云/Linux 服务器自动化环境部署。管理 Docker 环境、protobuf 脚手架、SSH 文件传输、图片处理、运动计时器 TUI。

## 工作流

1. **需求明确** — 不确定就问，不猜。有更简单做法要指出
2. **简单优先** — 只写当前问题所需最少代码，不做猜测式扩展。外科手术式修改，不碰无关代码
3. **TDD** (`github.com/stretchr/testify`) — 先写测试（红）→ 最少代码通过（绿）→ 重构

## 编码规范

- **接口**：单接口≤5方法，大接口用小接口聚合。实现名体现区别
- **Cobra**：rootCmd 统一全局 flag，子命令单文件。PreRunE 校验，RunE 业务，错误统一返回
- **输出**：stdout 正常输出，stderr 日志错误。支持 `--output json/table/yaml`，`--quiet`
- **文件/IO**：统一 `os + afero`、`filepath`。大文件流式，不一次加载全量
- **常量**：`config/const.go` — PascalCase，展开缩写（Def→Default），`{Scope}DirName` / `{Topic}Dir` 模式

## 工具使用

禁止大段的将代码加入上下文进行分析，可以使用 grep 等工具定位候选代码，使用 gopls mcp 工具精确锁定代码再加入上下问进行分析。

## 项目结构（概要）

```
seneschal/
├── main.go               # 入口
├── cmd/                   # Cobra 命令
├── config/                # 配置类型与加载（TOML/CSV/JSON）
├── internal/
│   ├── fsutil/            # 本地/远程统一文件系统（PathRef, alias:path）
│   ├── runner/            # 执行逻辑（SSH、Docker、环境管理器）
│   └── command/           # 命令业务实现（file操作、img、todo）
├── pkg/util/              # 工具函数（编辑器、表格、文件遍历）
└── ui/                    # 终端 UI（Bubble Tea 组件）
```

## 关键模式

- **配置**：`viper` + `file.ListFileNameWithExt()` 读取 TOML，`mapstructure.Decode` + `WriteConfigAs()` 写入
- **路径**：`fsutil.Parse("alias:/path")` → `PathRef`，`fsutil.GetFS()` 获取对应文件系统
- **代码注入**：`.proto` 文件内 `//rpc place` / `//message place` 探针，`internal/command/file/write.go`
- **环境管理**：`IEnvMgr` 接口（Check/Deploy），唯一实现 `EnvMgrDocker`
- **编辑器**：`pkg/util/editor.go:EditFile()` — `$VISUAL > $EDITOR > vi` 发现
- **模板**：Go `text/template`，配置在 `data/tpl/<name>/setting.toml`，输出到 `data/_gen/<name>/`

## 数据约定

| 配置 | 格式 | 目录 |
|------|------|------|
| SSH/Project/Env/Jenkins | TOML | `data/conf/<type>/` |
| 运动计划 | CSV | `data/conf/workout/` |
| Todo | JSON | `data/todo/` |
| 模板 | Go template | `data/tpl/<name>/template/` → `data/_gen/<name>/` |

# Seneschal - 个人向 CLI 工具箱

希望将 `Senschal` 打造成一个个人向 `CLI` 工具箱，尽可能涵盖常用功能：

1. `ssh` 客户端
2. `docker` 环境部署器
3. 配置生成器
4. `todo` 管理
5. 图片转化器
6. `protobuf` 文件工具
7. ...

## 用法

### 主机管理

```bash
# 添加 SSH 主机配置（支持密码/密钥认证）
seneschal host add myserver -u root --host 192.168.1.100 -p 22 --password mypassword
seneschal host add myserver -u root --host 192.168.1.100 --private ~/.ssh/id_rsa

# 列出已配置的主机
seneschal host list

# 上传文件/目录到一台或多台主机（别名逗号分隔）
seneschal host up server1,server2 ./config.toml /tmp/config.toml
seneschal host up server1 ./deploy/ /opt/app/          # → 远端 /opt/app/deploy/
seneschal host up server1 ./deploy/* /opt/app/         # → 直接拷贝内容到 /opt/app/

# 从主机下载文件/目录
seneschal host down server1 /var/log/app.log ./app.log
seneschal host down server1 /opt/data/ ./backup/

# 检查主机环境状态
seneschal host check server1,server2 docker

# 部署环境（Docker）到远程主机
seneschal host deploy server1,server2 docker
```

### 跨机文件传输（`cp`）

```bash
# 任意组合：本地↔远程、远程↔远程
seneschal cp ./local.txt server1:/remote/path/
seneschal cp server1:/src/file.txt server2:/dst/
```

### Jenkins

```bash
# 添加 Jenkins 配置
seneschal jenkins add myjenkins mytoken --host http://jenkins.example.com -u admin

# 添加使用自签证书的 Jenkins（跳过 TLS 验证）
seneschal jenkins add selfsigned <token> --host https://10.0.0.1:32005 --insecure

# 列出 Jenkins 实例的 Job
seneschal jenkins list          # 列出所有 Jenkins 配置
seneschal jenkins list myjenkins # 列出指定实例的 Job

# 从 XML 创建 Jenkins Job
seneschal jenkins create myjenkins --file job.xml --name my-job           # 单个 Job
seneschal jenkins create myjenkins --dir ./jobs/                          # 批量
seneschal jenkins create myjenkins --file job.xml --name my-job --overwrite # 覆盖
```

### 图片处理

```bash
# 图片转 ASCII 艺术字
seneschal img text photo.png -W 80 -H 40 -C   # ANSI 彩色输出
seneschal img text photo.png -W 80 -V          # 反转亮度

# GIF Sobel 边缘检测
seneschal img edge animation.gif
```

### Todo 管理

```bash
seneschal todo                    # 列出所有待办
seneschal todo add "完成文档"      # 添加待办
seneschal todo done 1              # 标记完成
seneschal todo del 1               # 删除待办
```

### 运动计时器

```bash
seneschal workout -l               # 列出可用运动计划
seneschal workout 腹肌撕裂者         # 开始运动（Bubble Tea TUI）
seneschal workout new myworkout    # 创建新的运动计划
```

### 环境配置

```bash
seneschal env     # 列出环境配置（Docker 版本、镜像等）
```

### Joynova 项目工具

```bash
# 列出项目配置
seneschal joy -l

# 注册 protobuf 接口（带 --lobby 同时注册大厅接口）
seneschal joy inter myproject --lobby usersvc:Login

# 模板管理
seneschal joy tpl                        # 列出可用模板
seneschal joy tpl exec mytpl             # 从模板生成文件
seneschal joy tpl exec mytpl -d ./out -s custom.toml -n myapp
```

## TODO

### Bug 修复

- [x] 渲染多文件夹下的同名模板时，所有生成结果使用了相同模板
      原因：使用ParseFiles/ParseGlob时，使用文件名作为模板唯一索引，如果出现同名模板时候会进行覆盖
      解决方案：使用完整路径作为模板索引，参考：

  ```golang
  // LoadTemplates 递归加载目录，模板名使用完整相对路径，杜绝重名冲突
  func LoadTemplates(root string) (*template.Template, error) {
    tpl := template.New("root") // 根容器，无实际渲染用途

    err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
      if err != nil {
        return err
      }
      if info.IsDir() {
        return nil
      }
      ext := filepath.Ext(path)
      if ext != ".tmpl" && ext != ".html" {
        return nil
      }

      // 关键：模板名 = 文件相对路径，全局唯一
      content, err := os.ReadFile(path)
      if err != nil {
        return err
      }
      _, err = tpl.New(path).Parse(string(content))
      return err
    })
    return tpl, err
  }

  func main() {
    tpl, _ := LoadTemplates("./templates")
    // 渲染 admin/list.tmpl，通过完整路径区分
    tpl.ExecuteTemplate(os.Stdout, "templates/admin/list.tmpl", nil)
    // 渲染 user/list.tmpl
    tpl.ExecuteTemplate(os.Stdout, "templates/user/list.tmpl", nil)
  }
  ```

  执行计划：
  - [x] 使用 `pkg/util/file` 中的工具函数筛选文件
  - [x] 使用完整相对路径作为唯一索引，避免模板覆盖
  - [x] 忽略空文件夹，无效需求

### 功能

- [x] 使用 slog 替换原日志
  - [x] 添加 trace_id
- [x] Jenkins 常用api接入（配置管理、Job列表）
- [x] Jenkins Job创建

### 优化

- [ ] 优化 joy inter 命令中 proto 文件编辑方式
      使用 `github.com/jhump/protoreflect/v2` 和 `github.com/bufbuild/protocompile` 更优雅的编辑proto文件
- [x] 实现 host up/down 文件上传下载命令
- [x] 实现 fsutil.CopyDir 目录递归拷贝
- [x] 重构文件传输功能，优化代码结构
- [x] 使用 `goph` 替换现有 `ssh` 客户端实现（放弃）。  
       测试发现 `goph` 中的封装并不适合本项目：
  1. 上传/下载文件仅支持单个文件，批量传输需要额外处理，且本项目已经实现了。
  2. ssh执行命令功能同样本项目已经实现

## License

本项目依据 [MIT](LICENSE) 协议开源。

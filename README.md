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

### 代理文件传输

```bash
# 列出 SSH 代理配置
seneschal agent list

# 上传文件到一台或多台代理（别名逗号分隔）
seneschal agent up server1,server2 ./config.toml /tmp/config.toml

# 上传目录
seneschal agent up server1 ./deploy/ /opt/app/

# 从代理下载文件
seneschal agent down server1 /var/log/app.log ./app.log

# 从代理下载目录
seneschal agent down server1 /opt/data/ ./backup/

# 代理间复制（任意组合：本地↔远程、远程↔远程）
seneschal agent cp ./local.txt server1:/remote/path/
seneschal agent cp server1:/src/file.txt server2:/dst/
```

### 代理环境管理

```bash
# 检查代理环境状态
seneschal agent check server1,server2 docker

# 部署环境
seneschal agent deploy server1,server2 docker
```

### 其他命令

```bash
seneschal img text photo.png -W 80 -C      # 图片转 ASCII
seneschal todo add "完成文档"               # 添加待办
seneschal workout -l                        # 运动计时器
```

## TODO

### 功能

- [] Jenkins 常用api接入

### 优化

- [x] 实现 agent up/down 文件上传下载命令
- [x] 实现 fsutil.CopyDir 目录递归拷贝
- [x] 重构文件传输功能，优化代码结构
- [x] 使用 `goph` 替换现有 `ssh` 客户端实现（放弃）。  
       测试发现 `goph` 中的封装并不适合本项目：
  1. 上传/下载文件仅支持单个文件，批量传输需要额外处理，且本项目已经实现了。
  2. ssh执行命令功能同样本项目已经实现

## License

本项目依据 [MIT](LICENSE) 协议开源。

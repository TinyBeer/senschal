# 自动环境安装工具
用于快速将云服务器改造为工作模板环境。
# TODO
## 交互方式
* [ ] manager workout config by csv file
* [ ] integrate image handler
* [ ] workout 
* [ ] 终端UI
* [ ] 交互式部署
## 文件
* [x] 文件拷贝 
* [x] 支持文件拷贝
    1. 如果是用于读取的文件，则进行预处理，递归获取文件信息包括路径大小等
## docker环境
### docker
* [x] docker 安装
  * [x] 使用英特网安装
  * [x] 使用deb安装包安装
* [x] 自动添加用户组
### docker镜像
* [x] docker 镜像加载
* [x] 优化镜像拉取 宿主机 docker拉取缺失的镜像
  * [x] 优化镜像拷贝  仅仅拷贝缺失镜像 
## 环境配置
* [x] 环境分组

## 输出美化
* [x] 美化输出 使用表格输出列表内容

# GHProxy

![pull](https://img.shields.io/docker/pulls/wjqserver/ghproxy.svg)![Docker Image Size (tag)](https://img.shields.io/docker/image-size/wjqserver/ghproxy/latest)[![Go Report Card](https://goreportcard.com/badge/github.com/WJQSERVER-STUDIO/ghproxy)](https://goreportcard.com/report/github.com/WJQSERVER-STUDIO/ghproxy)

使用Go实现的GHProxy,用于加速部分地区Github仓库的拉取,支持速率限制,用户鉴权,支持Docker部署

## 项目说明

### 项目特点

- ⚡ **基于 Go 语言实现，跨平台的同时提供高并发性能**
- 🌐 **使用字节旗下的 [HertZ](https://github.com/cloudwego/hertz) 作为 Web 框架**
- 📡 **使用 [Touka-HTTPC](https://github.com/satomitouka/touka-httpc) 作为 HTTP 客户端**
- 📥 **支持 Git clone、raw、releases 等文件拉取**
- 🎨 **支持多个前端主题**
- 🚫 **支持自定义黑名单/白名单**
- 🗄️ **支持 Git Clone 缓存（配合 [Smart-Git](https://github.com/WJQSERVER-STUDIO/smart-git)）**
- 🐳 **支持 Docker 部署**
- ⚡ **支持速率限制**
- 🔒 **支持用户鉴权**
- 🐚 **支持 shell 脚本嵌套加速**

### 项目相关

[DEMO](https://ghproxy.1888866.xyz)

[TG讨论群组](https://t.me/ghproxy_go)

[相关文章](https://blog.wjqserver.com/categories/my-program/)

### 使用示例

```
# 下载文件
https://ghproxy.1888866.xyz/raw.githubusercontent.com/WJQSERVER-STUDIO/tools-stable/main/tools-stable-ghproxy.sh
https://ghproxy.1888866.xyz/https://raw.githubusercontent.com/WJQSERVER-STUDIO/tools-stable/main/tools-stable-ghproxy.sh

# 克隆仓库
git clone https://ghproxy.1888866.xyz/github.com/WJQSERVER-STUDIO/ghproxy.git
git clone https://ghproxy.1888866.xyz/https://github.com/WJQSERVER-STUDIO/ghproxy.git
```

## 部署说明

可参考文章: https://blog.wjqserver.com/post/ghproxy-deploy-with-smart-git/

### Docker部署

- Docker-cli

```
docker run -p 7210:8080 -v ./ghproxy/log/run:/data/ghproxy/log -v ./ghproxy/log/caddy:/data/caddy/log -v ./ghproxy/config:/data/ghproxy/config  --restart always wjqserver/ghproxy
```

- Docker-Compose (建议使用)

    参看[docker-compose.yml](https://github.com/WJQSERVER-STUDIO/ghproxy/blob/main/docker/compose/docker-compose.yml)

### 二进制文件部署(不推荐)

一键部署脚本:

```bash
wget -O install.sh https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/deploy/install.sh && chmod +x install.sh &&./install.sh
```

Dev一键部署脚本:

```bash
wget -O install-dev.sh https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/dev/deploy/install-dev.sh && chmod +x install-dev.sh && ./install-dev.sh
```

## 配置说明

### 外部配置文件

本项目采用`config.toml`作为外部配置,默认配置如下
使用Docker部署时,慎重修改`config.toml`,以免造成不必要的麻烦

```toml
[server]
host = "0.0.0.0"  # 监听地址
port = 8080  # 监听端口
sizeLimit = 125 # 125MB
H2C = true # 是否开启H2C传输 
cors = "*" # "*"/"" -> "*" ; "nil" -> "" ; 除以上特殊情况, 会将值直接传入

[httpc]
mode = "auto" # "auto" or "advanced" HTTP客户端模式 自动/高级模式
maxIdleConns = 100 # only for advanced mode 仅用于高级模式
maxIdleConnsPerHost = 60 # only for advanced mode 仅用于高级模式
maxConnsPerHost = 0 # only for advanced mode 仅用于高级模式

[gitclone]
mode = "bypass" # bypass / cache 运行模式, cache模式依赖smart-git
smartGitAddr = "http://127.0.0.1:8080" # smart-git组件地址
ForceH2C = false # 强制使用H2C连接

[shell]
editor = false # 脚本嵌套加速

[pages]
mode = "internal" # "internal" or "external" 内部/外部 前端 默认内部
theme = "bootstrap" # "bootstrap" or "nebula" 内置主题
staticPath = "/data/www"  # 静态页面文件路径

[log]
logFilePath = "/data/ghproxy/log/ghproxy.log" # 日志文件路径
maxLogSize = 5 # MB 日志文件最大大小
level = "info"  # 日志级别 dump, debug, info, warn, error, none

[auth]
authMethod = "parameters" # 鉴权方式,支持parameters,header
authToken = "token"  # 用户鉴权Token
enabled = false  # 是否开启用户鉴权
ForceAllowApi = false # 在不开启Header鉴权的情况下允许api代理

[blacklist]
blacklistFile = "/data/ghproxy/config/blacklist.json"  # 黑名单文件路径
enabled = false  # 是否开启黑名单

[whitelist]
enabled = false  # 是否开启白名单
whitelistFile = "/data/ghproxy/config/whitelist.json"  # 白名单文件路径

[rateLimit]
enabled = false  # 是否开启速率限制
rateMethod = "total" # "ip" or "total" 速率限制方式
ratePerMinute = 180  # 每分钟限制请求数量
burst = 5  # 突发请求数量

[outbound]
enabled = false # 是否使用自定义代理出站
url = "socks5://127.0.0.1:1080" # "http://127.0.0.1:7890" 支持Socks5/HTTP(S)出站传输
```

### 黑名单配置

黑名单配置位于config/blacklist.json,格式如下:

```json
{
    "blacklist": [
      "test/test1",
      "example/repo2",
      "another/*"
      "another"
    ]
  }
```

### 白名单配置

白名单配置位于config/whitelist.json,格式如下:

```json
{
    "whitelist": [
      "test/test1",
      "example/repo2",
      "another/*"
      "another"
    ]
  }
```

### 前端页面

参看[GHProxy-Frontend](https://github.com/WJQSERVER-STUDIO/GHProxy-Frontend)

## 项目简史

**本项目是[WJQSERVER-STUDIO/ghproxy-go](https://github.com/WJQSERVER-STUDIO/ghproxy-go)的重构版本,实现了原项目原定功能的同时,进一步优化了性能**
关于此项目的详细开发过程,请参看Commit记录与[CHANGELOG.md](https://github.com/WJQSERVER-STUDIO/ghproxy/blob/main/CHANGELOG.md)

- v3.0.0 迁移到HertZ框架, 进一步提升效率
- v2.4.1 对路径匹配进行优化
- v2.0.0 对`proxy`核心模块进行了重构,大幅优化内存占用
- v1.0.0 迁移至本仓库,并再次重构内容实现
- v0.2.0 重构项目实现

## LICENSE

本项目使用WJQserver Studio License 2.0 [WJQserver Studio License 2.0](https://wjqserver-studio.github.io/LICENSE/LICENSE.html)

在v2.3.0之前, 本项目使用WJQserver Studio License 1.2

在v1.0.0版本之前,本项目继承于[WJQSERVER-STUDIO/ghproxy-go](https://github.com/WJQSERVER-STUDIO/ghproxy-go)的APACHE2.0 LICENSE VERSION

## 赞助

如果您觉得本项目对您有帮助,欢迎赞助支持,您的赞助将用于Demo服务器开支及开发者时间成本支出,感谢您的支持!

为爱发电,开源不易

爱发电: https://afdian.com/a/wjqserver

USDT(TRC20): `TNfSYG6F2vkiibd6J6mhhHNWDgWgNdF5hN`

### 捐赠列表

| 赞助人    |金额|
|--------|------|
| starry | 8 USDT (TRC20)   |

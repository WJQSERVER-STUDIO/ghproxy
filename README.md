# GHProxy

![pull](https://img.shields.io/docker/pulls/wjqserver/ghproxy.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/WJQSERVER-STUDIO/ghproxy)](https://goreportcard.com/report/github.com/WJQSERVER-STUDIO/ghproxy)

使用Go实现的GHProxy,用于加速部分地区Github仓库的拉取,支持速率限制,用户鉴权,支持Docker部署

[DEMO](https://ghproxy.1888866.xyz)

[TG讨论群组](https://t.me/ghproxy_go)

[版本更新介绍](https://blog.wjqserver.com/categories/my-program/)

## 项目说明

### 项目特点

- 基于Go语言实现,使用[Gin框架](https://github.com/gin-gonic/gin)与[req库](https://github.com/imroc/req)]
- 支持Git clone,raw,realeases等文件拉取
- 支持Docker部署
- 支持速率限制
- 支持用户鉴权
- 支持自定义黑名单/白名单
- 符合[RFC 7234](https://httpwg.org/specs/rfc7234.html)的HTTP Cache
- 使用Caddy作为Web Server
- 基于[WJQSERVER-STUDIO/golang-temp](https://github.com/WJQSERVER-STUDIO/golang-temp)模板构建,具有标准化的日志记录与构建流程

### 项目开发过程

**本项目是[WJQSERVER-STUDIO/ghproxy-go](https://github.com/WJQSERVER-STUDIO/ghproxy-go)的重构版本,实现了原项目原定功能的同时,进一步优化了性能**
关于此项目的详细开发过程,请参看Commit记录与[CHANGELOG.md](https://github.com/WJQSERVER-STUDIO/ghproxy/blob/main/CHANGELOG.md)

- V1.0.0 迁移至本仓库,并再次重构内容实现
- v0.2.0 重构项目实现

### LICENSE

本项目使用WSL LICENSE Version1.2 (WJQSERVER STUDIO LICENSE Version1.2)

在v1.0.0版本之前,本项目继承于[WJQSERVER-STUDIO/ghproxy-go](https://github.com/WJQSERVER-STUDIO/ghproxy-go)的APACHE2.0 LICENSE VERSION

## 使用示例

```
https://ghproxy.1888866.xyz/raw.githubusercontent.com/WJQSERVER-STUDIO/tools-stable/main/tools-stable-ghproxy.sh

git clone https://ghproxy.1888866.xyz/github.com/WJQSERVER-STUDIO/ghproxy.git
```

## 部署说明

### Docker部署

- Docker-cli

```
docker run -p 7210:80 -v ./ghproxy/log/run:/data/ghproxy/log -v ./ghproxy/log/caddy:/data/caddy/log -v ./ghproxy/config:/data/ghproxy/config  --restart always wjqserver/ghproxy
```

- Docker-Compose (建议使用)

    参看[docker-compose.yml](https://github.com/WJQSERVER-STUDIO/ghproxy/blob/main/docker/compose/docker-compose.yml)

### 二进制文件部署(不推荐)

一键部署脚本:

```bash
wget -O install.sh https://raw.githubusercontent.com/WJQSERVER-STUDIO/ghproxy/main/deploy/install.sh && chmod +x install.sh &&./install.sh
```

## 配置说明

### 外部配置文件

本项目采用`config.toml`作为外部配置,默认配置如下
使用Docker部署时,慎重修改`config.toml`,以免造成不必要的麻烦

```toml
[server]
host = "127.0.0.1"  # 监听地址
port = 8080  # 监听端口
sizeLimit = 125 # 125MB
enableH2C = "on"  # 是否开启H2C传输(latest和dev版本请开启) on/off

[pages]
enabled = false  # 是否开启内置静态页面(Docker版本请关闭此项)
staticPath = "/data/www"  # 静态页面文件路径

[log]
logFilePath = "/data/ghproxy/log/ghproxy.log" # 日志文件路径
maxLogSize = 5 # MB 日志文件最大大小

[cors]
enabled = true  # 是否开启跨域

[auth]
authMethod = "parameters" # 鉴权方式,支持parameters,header
authToken = "token"  # 用户鉴权Token
enabled = false  # 是否开启用户鉴权

[blacklist]
blacklistFile = "/data/ghproxy/config/blacklist.json"  # 黑名单文件路径
enabled = false  # 是否开启黑名单

[whitelist]
enabled = false  # 是否开启白名单
whitelistFile = "/data/ghproxy/config/whitelist.json"  # 白名单文件路径

[rateLimit]
enabled = false  # 是否开启速率限制
rateMrthod = "total" # "ip" or "total" 速率限制方式
ratePerMinute = 180  # 每分钟限制请求数量
burst = 5  # 突发请求数量
```

### 黑名单配置

黑名单配置位于config/blacklist.json,格式如下:

```json
{
    "blacklist": [
      "test/test1",
      "example/repo2",
      "another/*"
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
    ]
  }
```

### Caddy反代配置

```Caddyfile
example.com {
    reverse_proxy {
        to 127.0.0.1:7210
        header_up X-Real-IP {remote_host}	    
        header_up X-Real-IP {http.request.header.CF-Connecting-IP}
        header_up X-Forwarded-For {http.request.header.CF-Connecting-IP}
        header_up X-Forwarded-Proto {http.request.header.CF-Visitor}
    }
    encode zstd gzip    
}
```

### 前端页面

![ghproxy-demo.png](https://webp.wjqserver.com/ghproxy/ghproxy-demo-v1.7.0-mobile-night.png)

结语
---

本项目基于Go语言实现,使用Gin框架与req库
Docker镜像基于[WJQSERVER-STUDIO/caddy](https://github.com/WJQSERVER-STUDIO/caddy)
本项目使用WSL LICENSE Version1.2 (WJQSERVER STUDIO LICENSE Version1.2) 授权协议,请遵守相关条例。

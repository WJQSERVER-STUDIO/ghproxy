# GhProxy

![pull](https://img.shields.io/docker/pulls/wjqserver/ghproxy.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/WJQSERVER-STUDIO/ghproxy)](https://goreportcard.com/report/github.com/WJQSERVER-STUDIO/ghproxy)

使用Go实现的GHProxy,用于加速部分地区Github仓库的拉取,支持速率限制,用户鉴权,支持Docker部署

## 项目说明

### 项目特点

- 基于Go语言实现,使用[Gin框架](https://github.com/gin-gonic/gin)与[req库](https://github.com/imroc/req)]
- 支持Git clone,raw,realeases等文件拉取
- 支持Docker部署
- 支持速率限制
- 支持用户鉴权
- 使用Caddy作为Web Server
- 基于[WJQSERVER-STUDIO/golang-temp](https://github.com/WJQSERVER-STUDIO/golang-temp)模板构建,具有标准化的日志记录与构建流程

### 项目开发过程

**本项目是[WJQSERVER-STUDIO/ghproxy-go](https://github.com/WJQSERVER-STUDIO/ghproxy-go)的重构版本,实现了原项目原定功能的同时,进一步优化了性能**
本项目源于[WJQSERVER-STUDIO/ghproxy-go](https://github.com/WJQSERVER-STUDIO/ghproxy-go)与[WJQSERVER/ghproxy-go-0RTT](https://github.com/WJQSERVER/ghproxy-go-0RTT)两个项目,前者带来了实现框架与资源,后者带来了解决Git clone问题的办法,使得本项目从net/http标准库切换至Gin框架,已解决此困扰已久的问题,在此基础上,本项目进一步优化了性能,并添加了用户鉴权功能,使得部署更加安全可靠。
关于此项目的详细开发过程,请参看Commit记录与[CHANGELOG.md](https://github.com/WJQSERVER-STUDIO/ghproxy/blob/main/CHANGELOG.md)

- V1.0.0 迁移至本仓库,并再次重构内容实现
- v0.2.0 重构项目实现,Git clone的实现完全自主化

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
docker run -p 7210:80 -v ./ghproxy/log/run:/data/ghproxy/log -v ./ghproxy/log/caddy:/data/caddy/log --restart always wjqserver/ghproxy
```

- Docker-Compose

    参看[docker-compose.yml](https://github.com/WJQSERVER-STUDIO/ghproxy/blob/main/docker/compose/docker-compose.yml)

### 外部配置文件

本项目采用config.yaml作为外部配置,默认配置如下
使用Docker部署时,慎重修改config.yaml,以免造成不必要的麻烦

```
port: 8080  # 监听端口
host: "127.0.0.1"  # 监听地址
sizelimit: 131072000 # 125MB
logfilepath: "/data/ghproxy/log/ghproxy.log"  # 日志文件路径
CorsAllowOrigins: true  # 是否允许跨域请求
auth: true  # 是否开启鉴权
authtoken: "test"  # 鉴权token
```

### Caddy反代配置

```
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

## TODO & DEV

### TODO

- [x] 允许更多参数通过config结构传入
- [x] 改进程序效率
- [x] 用户鉴权

### DEV

- [x] Docker Pull 代理

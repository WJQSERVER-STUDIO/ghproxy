# GHProxy

![GitHub Release](https://img.shields.io/github/v/release/WJQSERVER-STUDIO/ghproxy?display_name=tag&style=flat)
![pull](https://img.shields.io/docker/pulls/wjqserver/ghproxy.svg)
![Docker Image Size (tag)](https://img.shields.io/docker/image-size/wjqserver/ghproxy/latest)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/WJQSERVER-STUDIO/ghproxy)
[![Go Report Card](https://goreportcard.com/badge/github.com/WJQSERVER-STUDIO/ghproxy)](https://goreportcard.com/report/github.com/WJQSERVER-STUDIO/ghproxy)

一个基于Go的高性能Github资源代理程序, 同时支持Docker镜像代理与脚本嵌套加速等多种功能

## 项目说明

### 项目特点

- ⚡ **基于 Go 语言实现，跨平台的同时提供高并发性能**
- 🌐 **使用自有[Touka框架](https://github.com/infinite-iroha/touka)作为 HTTP服务端框架**
- 📡 **使用 [Touka-HTTPC](https://github.com/WJQSERVER-STUDIO/httpc) 作为 HTTP 客户端**
- 📥 **支持 Git clone、raw、releases 等文件拉取**
- 🐳 **支持反代Docker, GHCR等镜像仓库**
- 🎨 **支持多个前端主题**
- 🚫 **支持自定义黑名单/白名单**
- 🗄️ **支持 Git Clone 缓存（配合 [Smart-Git](https://github.com/WJQSERVER-STUDIO/smart-git)）**
- 🐳 **支持自托管与Docker容器化部署**
- ⚡ **支持速率限制**
- ⚡ **支持带宽速率限制**
- 🔒 **支持用户鉴权**
- 🐚 **支持 shell 脚本多层嵌套加速**

### 项目相关

[DEMO](https://ghproxy.1888866.xyz)

[TG讨论群组](https://t.me/ghproxy_go)

[GHProxy项目文档](https://wjqserver-docs.pages.dev/docs/ghproxy/) 感谢 [@redbunnys](https://github.com/redbunnys)的维护

[相关文章](https://blog.wjqserver.com/categories/my-program/)

### 使用示例

```bash 
# 下载文件
https://ghproxy.1888866.xyz/raw.githubusercontent.com/WJQSERVER-STUDIO/tools-stable/main/tools-stable-ghproxy.sh
https://ghproxy.1888866.xyz/https://raw.githubusercontent.com/WJQSERVER-STUDIO/tools-stable/main/tools-stable-ghproxy.sh

# 克隆仓库
git clone https://ghproxy.1888866.xyz/github.com/WJQSERVER-STUDIO/ghproxy.git
git clone https://ghproxy.1888866.xyz/https://github.com/WJQSERVER-STUDIO/ghproxy.git

# Docker(OCI) 代理
docker pull gh.example.com/wjqserver/ghproxy
docker pull gh.example.com/adguard/adguardhome

docker pull gh.example.com/docker.io/wjqserver/ghproxy
docker pull gh.example.com/docker.io/adguard/adguardhome

docker pull gh.example.com/ghcr.io/openfaas/queue-worker 
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

参看[项目文档](https://github.com/WJQSERVER-STUDIO/ghproxy/blob/main/docs/config.md)

### 前端页面

参看[GHProxy-Frontend](https://github.com/WJQSERVER-STUDIO/GHProxy-Frontend)

## 文档

*   [GHProxy项目文档](https://wjqserver-docs.pages.dev/docs/ghproxy/) 感谢 [@redbunnys](https://github.com/redbunnys)的维护

*   [![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/WJQSERVER-STUDIO/ghproxy) 可供参考, AI生成存在幻觉, 不完全可靠, 请注意辨别

## LICENSE

v3.5.2开始, 本项目使用 [WJQserver Studio License 2.1](https://wjqserver-studio.github.io/LICENSE/LICENSE.html) 和 [Mozilla Public License Version 2.0](https://mozilla.org/MPL/2.0/) 双重许可, 您可从中选择一个使用

前端位于单独仓库中, 且各个主题均存在各自的许可证, 本项目许可证并不包括前端

在v2.3.0之前, 本项目使用WJQserver Studio License 1.2

在v1.0.0版本之前,本项目继承于[WJQSERVER-STUDIO/ghproxy-go](https://github.com/WJQSERVER-STUDIO/ghproxy-go)的APACHE2.0 LICENSE VERSION

## 赞助

如果您觉得本项目对您有帮助,欢迎赞助支持,您的赞助将用于Demo服务器开支及开发者时间成本支出,感谢您的支持!

USDT(TRC20): `TNfSYG6F2vkiibd6J6mhhHNWDgWgNdF5hN`

### 捐赠列表

| 赞助人    |金额|
|--------|------|
| starry | 8 USDT (TRC20)   |

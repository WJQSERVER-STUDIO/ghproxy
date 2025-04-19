# ghproxy 用户配置文档

`ghproxy` 的配置主要通过修改 `config` 目录下的 `config.toml`、`blacklist.json` 和 `whitelist.json` 文件来实现。本文档将详细介绍这些配置文件的作用以及用户可以自定义的配置选项。

## `config.toml` - 主配置文件

`config.toml` 是 `ghproxy` 的主配置文件，采用 TOML 格式。您可以通过修改此文件来定制 `ghproxy` 的各项功能，例如服务器端口、连接设置、Git 克隆模式、日志级别、认证方式、黑白名单以及限速策略等。

以下是 `config.toml` 文件的详细配置项说明：

```toml name=config/config.toml
[server]
host = "0.0.0.0"
port = 8080
netlib = "netpoll" # "netpoll" / "std" "standard" "net/http" "net"
sizeLimit = 125 # MB
memLimit = 0 # MB
H2C = true
cors = "*" # "*"/"" -> "*" ; "nil" -> "" ;
debug = false

[httpc]
mode = "auto" # "auto" or "advanced"
maxIdleConns = 100 # only for advanced mode
maxIdleConnsPerHost = 60 # only for advanced mode
maxConnsPerHost = 0 # only for advanced mode

[gitclone]
mode = "bypass" # bypass / cache
smartGitAddr = "http://127.0.0.1:8080"
ForceH2C = false

[shell]
editor = false
rewriteAPI = false

[pages]
mode = "internal" # "internal" or "external"
theme = "bootstrap" # "bootstrap" or "nebula"
staticDir = "/data/www"

[log]
logFilePath = "/data/ghproxy/log/ghproxy.log"
maxLogSize = 5 # MB
level = "info" # dump, debug, info, warn, error, none
hertzLogPath = "/data/ghproxy/log/hertz.log"

[auth]
method = "parameters" # "header" or "parameters"
token = "token"
key = ""
enabled = false
passThrough = false
ForceAllowApi = false

[blacklist]
blacklistFile = "/data/ghproxy/config/blacklist.json"
enabled = false

[whitelist]
enabled = false
whitelistFile = "/data/ghproxy/config/whitelist.json"

[rateLimit]
enabled = false
rateMethod = "total" # "ip" or "total"
ratePerMinute = 180
burst = 5

[outbound]
enabled = false
url = "socks5://127.0.0.1:1080" # "http://127.0.0.1:7890"
```

### 配置项详细说明

*   **`[server]` - 服务器配置**

    *   `host`:  监听地址。
        *   类型: 字符串 (`string`)
        *   默认值: `"0.0.0.0"` (监听所有)
        *   说明:  设置 `ghproxy` 监听的网络地址。通常设置为 `"0.0.0.0"` 以监听所有可用的网络接口。
    *   `port`:  监听端口。
        *   类型: 整数 (`int`)
        *   默认值: `8080`
        *   说明:  设置 `ghproxy` 监听的端口号。
    *   `netlib`: 底层网络库。
        *   类型: 字符串 (`string`)
        *   默认值: `""` (HertZ默认处置)
        *   说明: `"std"` `"standard"` `"net/http"` `"net"` 均会被设置为go标准库`net/http`, 设置为`"netpoll"`或`""`会由`HertZ`默认逻辑处理
    *   `sizeLimit`: 请求体大小限制。
        *   类型: 整数 (`int`)
        *   默认值: `125` (MB)
        *   说明:  限制允许接收的请求体最大大小，单位为 MB。用于防止过大的请求导致服务压力过大。
    *   `memLimit`:  `runtime`内存限制
        *   类型: 整数 (`int64`)
        *   默认值: `0` (不传入)
        *   说明: 给`runtime`的指标, 让gc行为更高效
    *   `H2C`:  是否启用 H2C (HTTP/2 Cleartext) 传输。
        *   类型: 布尔值 (`bool`)
        *   默认值: `true` (启用)
        *   说明:  启用后，允许客户端使用 HTTP/2 协议进行无加密传输，提升性能。
    *   `cors`:  CORS (跨域资源共享) 设置。
        *   类型: 字符串 (`string`)
        *   默认值: `"*"` (允许所有来源)
        *   可选值:
            *   `""` 或`"*"`: 允许所有来源跨域访问。
            *   `"nil"`:  禁用 CORS。
            *   具体的域名:  例如 `"https://example.com"`，只允许来自指定域名的跨域请求。
        *   说明:  配置 CORS 策略，用于控制哪些域名可以跨域访问 `ghproxy` 服务。
    *   `debug`:  是否启用调试模式。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后，`ghproxy` 会输出更详细的日志信息，用于开发和调试。

*   **`[httpc]` - HTTP 客户端配置**

    *   `mode`:  HTTP 客户端模式。
        *   类型: 字符串 (`string`)
        *   默认值: `"auto"` (自动模式)
        *   可选值:
            *   `"auto"`:  自动模式，使用默认的 HTTP 客户端配置，适用于大多数场景。
            *   `"advanced"`: 高级模式，允许自定义连接池参数，可以更精细地控制 HTTP 客户端的行为。
        *   说明:  选择 HTTP 客户端的运行模式。
    *   `maxIdleConns`:  最大空闲连接数 (仅在高级模式下生效)。
        *   类型: 整数 (`int`)
        *   默认值: `100`
        *   说明:  设置 HTTP 客户端连接池中保持的最大空闲连接数。
    *   `maxIdleConnsPerHost`:  每个主机最大空闲连接数 (仅在高级模式下生效)。
        *   类型: 整数 (`int`)
        *   默认值: `60`
        *   说明:  设置 HTTP 客户端连接池中，每个主机允许保持的最大空闲连接数。
    *   `maxConnsPerHost`:  每个主机最大连接数 (仅在高级模式下生效)。
        *   类型: 整数 (`int`)
        *   默认值: `0` (不限制)
        *   说明:  设置 HTTP 客户端连接池中，每个主机允许建立的最大连接数。设置为 `0` 表示不限制。

*   **`[gitclone]` - Git 克隆配置**

    *   `mode`:  Git 克隆模式。
        *   类型: 字符串 (`string`)
        *   默认值: `"bypass"` (绕过模式)
        *   可选值:
            *   `"bypass"`:  绕过模式，直接克隆 GitHub 仓库，不使用任何缓存加速。
            *   `"cache"`:  缓存模式，使用智能 Git 服务加速克隆，需要配置 `smartGitAddr`。
        *   说明:  选择 Git 克隆的模式。
    *   `smartGitAddr`:  智能 Git 服务地址 (仅在缓存模式下生效)。
        *   类型: 字符串 (`string`)
        *   默认值: `"http://127.0.0.1:8080"`
        *   说明:  当 `mode` 设置为 `"cache"` 时，需要配置智能 Git 服务的地址，用于加速 Git 克隆。
    *   `ForceH2C`:  是否强制使用 H2C 连接到智能 Git 服务。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (不强制)
        *   说明:  如果智能 Git 服务支持 H2C，可以设置为 `true` 以强制使用 H2C 连接，提升性能。

*   **`[shell]` - Shell 嵌套加速功能配置**

    *   `editor`:  是否启用编辑(嵌套加速)功能。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后, 会修改`.sh`文件内容以实现嵌套加速
    *   `rewriteAPI`:  是否重写 API 地址。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后，`ghproxy` 会重写脚本内的Github API地址。

*   **`[pages]` - Pages 服务配置**

    *   `mode`:  Pages 服务模式。
        *   类型: 字符串 (`string`)
        *   默认值: `"internal"` (内置 Pages 服务)
        *   可选值:
            *   `"internal"`:  使用 `ghproxy` 内置的 Pages 服务。
            *   `"external"`:  使用外部 Pages 位置。
        *   说明:  选择 Pages 服务的运行模式。
    *   `theme`:  Pages 主题。
        *   类型: 字符串 (`string`)
        *   默认值: `"bootstrap"`
        *   可选值: 参看[GHProxy项目前端仓库](https://github.com/WJQSERVER-STUDIO/GHProxy-Frontend)
        *   说明:  设置内置 Pages 服务使用的主题。
    *   `staticDir`:  静态文件目录。
        *   类型: 字符串 (`string`)
        *   默认值: `"/data/www"`
        *   说明:  指定外置 Pages 服务使用的静态文件目录。

*   **`[log]` - 日志配置**

    *   `logFilePath`:  日志文件路径。
        *   类型: 字符串 (`string`)
        *   默认值: `"/data/ghproxy/log/ghproxy.log"`
        *   说明:  设置 `ghproxy` 日志文件的存储路径。
    *   `maxLogSize`:  最大日志文件大小。
        *   类型: 整数 (`int`)
        *   默认值: `5` (MB)
        *   说明:  设置单个日志文件的最大大小，单位为 MB。当日志文件大小超过此限制时，会进行日志轮转。
    *   `level`:  日志级别。
        *   类型: 字符串 (`string`)
        *   默认值: `"info"`
        *   可选值: `"dump"`, `"debug"`, `"info"`, `"warn"`, `"error"`, `"none"`
        *   说明:  设置日志输出的级别。级别越高，输出的日志信息越少。
            *   `"dump"`:  输出所有日志，包括最详细的调试信息。
            *   `"debug"`:  输出调试信息、信息、警告和错误日志。
            *   `"info"`:   输出信息、警告和错误日志。
            *   `"warn"`:   输出警告和错误日志。
            *   `"error"`:  仅输出错误日志。
            *   `"none"`:   禁用所有日志输出。
    *   `hertzLogPath`:  `HertZ`日志文件路径。
        *   类型: 字符串 (`string`)
        *   默认值: `"/data/ghproxy/log/hertz.log"`
        *   说明:  设置 `HertZ` 日志文件的存储路径。

*   **`[auth]` - 认证配置**

    *   `enabled`:  是否启用认证。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后，需要提供正确的认证信息才能访问 `ghproxy` 服务。
    *   `method`:  认证方法。
        *   类型: 字符串 (`string`)
        *   默认值: `"parameters"` (URL 参数)
        *   可选值: `"header"` 或 `"parameters"`
            *   `"header"`:  通过请求头 `GH-Auth` 或自定义请求头 (通过 `key` 配置) 传递认证 Token。
            *   `"parameters"`: 通过 URL 参数 `auth_token` 或自定义 URL 参数名 (通过 `Key` 配置) 传递认证 Token。
        *   说明:  选择认证信息传递的方式。
    *   `key`:  自定义认证 Key。
        *   类型: 字符串 (`string`)
        *   默认值: `""` (空字符串，使用默认的 `GH-Auth` 请求头或 `auth_token` URL 参数名)
        *   说明:  可以自定义认证时使用的请求头名称或 URL 参数名。如果为空，则使用默认的 `GH-Auth` 请求头或 `auth_token` URL 参数名。
    *   `token`:  认证 Token。
        *   类型: 字符串 (`string`)
        *   默认值: `"token"`
        *   说明:  设置认证时需要提供的 Token 值。
    *   `passThrough`:  是否认证参数透穿到Github。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (不允许)
        *   说明:  如果设置为 `true`，相关参数会被透穿到Github。
    *   `ForceAllowApi`:  是否强制允许 API 访问。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (不强制允许)
        *   说明:  如果设置为 `true`，则强制允许对 GitHub API 的访问，即使未启用认证或认证失败。

*   **`[blacklist]` - 黑名单配置**

    *   `enabled`:  是否启用黑名单。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后，`ghproxy` 将根据 `blacklist.json` 文件中的规则阻止对特定用户或仓库的访问。
    *   `blacklistFile`:  黑名单文件路径。
        *   类型: 字符串 (`string`)
        *   默认值: `"/data/ghproxy/config/blacklist.json"`
        *   说明:  指定黑名单配置文件的路径。

*   **`[whitelist]` - 白名单配置**

    *   `enabled`:  是否启用白名单。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后，`ghproxy` 将只允许访问 `whitelist.json` 文件中规则指定的用户或仓库。白名单的优先级高于黑名单。
    *   `whitelistFile`:  白名单文件路径。
        *   类型: 字符串 (`string`)
        *   默认值: `"/data/ghproxy/config/whitelist.json"`
        *   说明:  指定白名单配置文件的路径。

*   **`[rateLimit]` - 限速配置**

    *   `enabled`:  是否启用限速。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后，`ghproxy` 将根据配置的策略限制请求速率，防止服务被滥用。
    *   `rateMethod`:  限速方法。
        *   类型: 字符串 (`string`)
        *   默认值: `"total"` (全局限速)
        *   可选值: `"ip"` 或 `"total"`
            *   `"ip"`:  基于客户端 IP 地址进行限速，每个 IP 地址都有独立的速率限制。
            *   `"total"`: 全局限速，所有客户端共享同一个速率限制。
        *   说明:  选择限速的策略。
    *   `ratePerMinute`:  每分钟允许的请求数。
        *   类型: 整数 (`int`)
        *   默认值: `180`
        *   说明:  设置每分钟允许通过的最大请求数。
    *   `burst`:  突发请求数。
        *   类型: 整数 (`int`)
        *   默认值: `5`
        *   说明:  允许在短时间内超过 `ratePerMinute` 的突发请求数。

*   **`[outbound]` - 出站代理配置**

    *   `enabled`:  是否启用出站代理。
        *   类型: 布尔值 (`bool`)
        *   默认值: `false` (禁用)
        *   说明:  启用后，`ghproxy` 将通过配置的代理服务器转发所有出站请求。
    *   `url`:  出站代理 URL。
        *   类型: 字符串 (`string`)
        *   默认值: `"socks5://127.0.0.1:1080"`
        *   支持协议: `socks5://` 和 `http://`
        *   说明:  设置出站代理服务器的 URL。支持 SOCKS5 和 HTTP 代理协议。

## `blacklist.json` - 黑名单配置

`blacklist.json` 文件用于配置黑名单规则，阻止对特定用户或仓库的访问。

```json name=config/blacklist.json
{
  "blacklist": [
    "eviluser",
    "spamuser/bad-repo",
    "malwareuser/*"
  ]
}
```

### 黑名单规则说明

*   `blacklist`:  一个 JSON 数组，包含黑名单规则，每条规则为一个字符串。
    *   **用户名**: 例如 `"eviluser"`，阻止所有名为 `eviluser` 的用户的访问。
    *   **仓库名**: 例如 `"spamuser/bad-repo"`，阻止访问 `spamuser` 用户下的 `bad-repo` 仓库。
    *   **通配符**: 例如 `"malwareuser/*"`，使用 `*` 通配符，阻止访问 `malwareuser` 用户下的所有仓库。
    *   **缩略写法**: 例如 `"example"`, 等同于 `"example/*"`， 允许访问 `example` 用户下的所有仓库。

## `whitelist.json` - 白名单配置

`whitelist.json` 文件用于配置白名单规则，只允许访问白名单中指定的用户或仓库。白名单的优先级高于黑名单，如果一个请求同时匹配黑名单和白名单，则白名单生效，请求将被允许。

```json name=config/whitelist.json
{
  "whitelist": [
    "white/list",
    "white/test1",
    "example/*",
    "example"
  ]
}
```

### 白名单规则说明

*   `whitelist`:  一个 JSON 数组，包含白名单规则，每条规则为一个字符串。
    *   **仓库名**: 例如 `"white/list"`，允许访问 `white` 用户下的 `list` 仓库。
    *   **仓库名**: 例如 `"white/test1"`，允许访问 `white` 用户下的 `test1` 仓库。
    *   **通配符**: 例如 `"example/*"`，使用 `*` 通配符，允许访问 `example` 用户下的所有仓库。
    *   **缩略写法**: 例如 `"example"`, 等同于 `"example/*"`， 允许访问 `example` 用户下的所有仓库。

---
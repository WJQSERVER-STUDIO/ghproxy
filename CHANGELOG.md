# 更新日志

24w12b
---
- PRE-RELEASE: 此版本是v1.4.1的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 尝试引入Alpine Linux作为基础镜像,减少镜像体积

24w12a
---
- PRE-RELEASE: 此版本是v1.4.1的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- FIX: 修正部分参数错误
- CHANGE: CGO_ENABLED=0

v1.4.0
---
- CHANGE: 优化代码结构,提升性能
- ADD: 新增auth子模块whitelist.go,支持白名单功能
- ADD: 新增whitelist.json文件,用于配置白名单
- CHANGE&ADD: 在config.yaml文件中新增白名单配置块
- FIX: 由于临时加入且未在原开发路线上计划的白名单功能,导致函数命名冲突,在此修复blacklist.go的函数命名问题
- FIX: 修复黑/白名单是否生效相关问题

24w11b
---
- PRE-RELEASE: 此版本是v1.4.0的预发布版本,请勿在生产环境中使用
- FIX: 修复黑/白名单是否生效相关问题

24w11a
---
- PRE-RELEASE: 此版本是v1.4.0的预发布版本,请勿在生产环境中使用
- **ANNOUNCE**: 自此版本起,DEV版本号格式进行修改,小版本号不再仅限于a/b,而是采用字母表顺序进行排列,此修改将带来一个重要改变,正式版前的预发布版本的数字版本号将会统一，以便于版本管理与发布管理
- CHANGE: 优化代码结构,提升性能
- ADD: 新增auth子模块whitelist.go,支持白名单功能
- ADD: 新增whitelist.json文件,用于配置白名单
- FIX: 由于临时加入且未在原开发路线上计划的白名单功能,导致函数命名冲突,在此修复blacklist.go的函数命名问题

v1.3.1
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: 剃刀计划,减少多余日志输出
- CHANGE: 调整缓存参数

24w10a
---
- PRE-RELEASE: 此版本是v1.3.1的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 剃刀计划,减少多余日志输出
- CHANGE: 调整缓存参数

v1.3.0
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: 优化黑名单功能,提升稳定性
- CHANGE: 剃刀计划,减少多余日志输出
- ADD： 新增auth子模块blacklist.go,支持黑名单功能
- ADD: 新增blacklist.json文件,用于配置黑名单
- CHANGE: config.yaml文件格式修改,使其具备更好的可读性
- WARNING: 此版本为大版本更新,配置文件重构,此版本不再向前兼容,请注意备份文件并重新部署

24w09b
---
- PRE-RELEASE: 此版本是v1.3.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 修正配置,提升稳定性
- WARNING: 此版本配置文件重构,此版本不再向前兼容,请注意备份文件并重新部署

24w09a
---
- PRE-RELEASE: 此版本是v1.3.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 优化黑名单功能,提升稳定性
- CHANGE&ADD: 新增auth子模块blacklist.go
- CHANGE: 黑名单转为使用json文件存储,便于程序处理
- WARNING: 此版本配置文件重构,此版本不再向前兼容,请注意备份文件并重新部署

24w08b
---
- PRE-RELEASE: 此版本是v1.3.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- ADD & CHANGE: 新增仓库黑名单功能,改进Auth模块
- ADD: 新增blacklist.yaml文件,用于配置仓库黑名单
- CHANGE: 大幅度修改Config包,使其更加模块化
- CHANGE: 与Config包同步修改config.yaml文件(不向前兼容)
- CHANGE: 修改config.yaml文件的格式,使其具备更好的可读性
- WARNING: 此版本配置文件重构,此版本不再向前兼容,请注意备份文件并重新部署

v1.2.0
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: 同步更新logger模块，与golang-temp项目定义的开发规范保持一致
- ADD: 新增日志翻转功能

24w08a
---
- PRE-RELEASE: 此版本是v1.2.0的预发布版本,请勿在生产环境中使用
- CHANGE: 同步更新logger模块，与golang-temp项目定义的开发规范保持一致
- ADD: 新增日志翻转功能

v1.1.1
---
- CHANGE: 修改部分代码，与golang-temp项目定义的开发规范保持一致
- CHANGE: 更新Go版本至v1.23.2
- CHANGE: 跟随Caddy更新,修改Caddyfile配置

24w07b
---
- PRE-RELEASE: 此版本是v1.1.1的预发布版本,请勿在生产环境中使用
- CHANGE: 修改部分代码，与golang-temp项目定义的开发规范保持一致
- CHANGE: 更新Go版本至v1.23.2
- CHANGE: 跟随Caddy更新,修改Caddyfile配置

24w07a
---
- PRE-RELEASE: 此版本是v1.1.1的预发布版本,请勿在生产环境中使用
- CHANGE: 修改部分代码，与golang-temp项目定义的开发规范保持一致
- CHANGE: 更新Go版本至v1.23.2

v1.1.0
---
- CHANGE: 优化代码结构,对main函数进行模块化,提升可读性
- CHANGE: Docker代理功能移至DEV版本内,保证稳定性
- ADD&CHANGE: 增加Auth(用户鉴权)模块,并改进其的错误处理与日志记录
- CHANGE: 日志模块引入goroutine协程,提升性能
- ADD: 将主要实现分离,作为Proxy模块,并优化代码结构
- ADD: 新增安全政策

24w06b
---
- PRE-RELEASE: 此版本是v1.1.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,对main函数进行模块化,提升可读性
- CHANGE: Docker代理功能移至DEV版本内,保证稳定性
- ADD&CHANGE: 增加Auth(用户鉴权)模块,并改进其的错误处理与日志记录
- CHANGE: 日志模块引入goroutine协程,提升性能 (实验性功能)
- ADD: 将主要实现分离,作为Proxy模块,并优化代码结构
- ADD: 新增安全政策

v1.0.0
---
- **ANNOUNCE**: 项目正式发布, 并迁移至[WJQSERVER-STUDIO/ghproxy](https://github.com/WJQSERVER-STUDIO/ghproxy)，由Apache License Version 2.0转为WJQserver Studio License 请注意相关条例变更
- CHANGE: 项目正式发布, 并迁移至[WJQSERVER-STUDIO/ghproxy](https://github.com/WJQSERVER-STUDIO/ghproxy)
- CHANGE: 再次重构代码,优化性能,提升稳定性
- CHANGE: 使用golang-temp项目作为底层构建,标准化日志与配置模块
- CHANGE: 从原项目的Apache License Version 2.0迁移至WJQserver Studio License
  
24w06a
---
- PRE-RELEASE: 此版本是v1.0.0的预发布版本,请勿在生产环境中使用
- CHANGE: 与v1.0.0版本同步

v0.2.0
---
底层核心代码重写,彻底代表着项目进入自主可控的开发阶段,彻底脱离原有实现
- ADD: 增加多处日志记录,便于审计与排障
- CHANGE: 优化代码结构,进一步模块化,同时提升性能
- ADD： 使用req库重构代码,提升请求伪装能力,尽可能bypass反爬机制

24w05b
---
- PRE-RELEASE: 此版本是v0.2.0的预发布版本,请勿在生产环境中使用
- CHANGE: 重命名proxychrome函数
- ADD: 增加多处日志记录,便于审计与排障

24w05a
---
- PRE-RELEASE: 此版本是v0.2.0的预发布版本,请勿在生产环境中使用
- FIX： 修正上一版本的req请求未继承请求方式的问题
- CHANGE: 优化代码结构,进一步模块化,同时提升性能

24w04b
---
- PRE-RELEASE: 此版本是v0.2.0的预发布版本,请勿在生产环境中使用
- CHANGE: 更换Docker基础镜像为daily版本
- ADD： 新增使用req库实现代理请求,使用chrome TLS指纹发起请求

24w04a
---
- PRE-RELEASE: 此版本是v0.2.0的预发布版本,请勿在生产环境中使用
- CHANGE: 调整程序结构,使用init函数初始化配置,并优化代码结构

v0.1.7
---
- CHANGE: 合入上游(wjqserver/caddy:latest)安全更新, 增强镜像安全性

24w03b
---
- PRE-RELEASE: 此版本是v0.1.7的预发布版本,请勿在生产环境中使用
- CHANGE: 合入上游(wjqserver/caddy:latest)安全更新, 增强镜像安全性

v0.1.6
---
- ADD: 新增跨域配置选项
- CHANGE: 更新UA标识

24w03a
---
- PRE-RELEASE: 此版本是v0.1.6的预发布版本,请勿在生产环境中使用
- CHANGE: 改进Docker代理相关Caddy配置
- ADD: 新增跨域配置选项

v0.1.5
---
- CHANGE: 更新Go版本至v1.23.1
- CHANGE: 优化代码行为

24w02b
---
- PRE-RELEASE: 此版本是v0.1.5的预发布版本,请勿在生产环境中使用
- ADD: 新增Docker代理 (未并入正式版)

24w02a
---
- PRE-RELEASE: 此版本是v0.1.5的预发布版本,请勿在生产环境中使用
- CHANGE: 更新Go版本至v1.23.1
- CHANGE: 优化代码行为

v0.1.4
---
- ADD: 新增外部文件配置功能
- ADD: 新增日志功能
- CHANGE: 优化代码结构,提升性能
- CHANGE: 改进前端界面,加入页脚

24w01b
---
- PRE-RELEASE: 此版本是v0.1.4的预发布版本,请勿在生产环境中使用
- ADD: 新增外部文件配置功能
- ADD: 新增日志功能
- CHANGE: 优化代码结构,提升性能
- CHANGE: 改进前端界面,加入页脚

v0.1.3
---
- **ANNOUNCE**: 开始自行维护项目,脱离上游更新
- CHANGE: 改进已有实现,增强程序稳定性

24w01a
---
- PRE-RELEASE: 此版本是v0.1.3的预发布版本,请勿在生产环境中使用
- **ANNOUNCE**: 首个DEV版本发布
- CHANGE: 同步更新

v0.1.2
---
- ADD: 新增项目介绍
- CHANGE: 限制默认文件大小限制到256M

v0.1.1
---
- ADD: Apache License Version 2.0
- FIX: 改进部分代码逻辑
- CHANGE: 将Go升级至v1.23.0

v0.1.0
---
项目的第一个版本
- ADD: 实现速率限制
- ADD: 实现符合[RFC 7234](https://httpwg.org/specs/rfc7234.html)的HTTP缓存机制
- ADD: 实现action编译
- ADD: 实现Docker部署
- INFO: 使用Caddy作为Web服务器，通过Caddy实现了缓存与速率限制

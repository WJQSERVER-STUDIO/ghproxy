# 更新日志

3.2.0 - 2025-04-27
---
- CHANGE: 加入`ghcr`和`dockerhub`反代功能
- FIX: 修复在`HertZ`路由匹配器下与用户名相关功能异常的问题

25w31a - 2025-04-27
---
- PRE-RELEASE: 此版本是v3.2.0预发布版本,请勿在生产环境中使用;
- CHANGE: 加入`ghcr`和`dockerhub`反代功能
- FIX: 修复在`HertZ`路由匹配器下与用户名相关功能异常的问题

3.1.0 - 2025-04-24
---
- CHANGE: 对标准url使用`HertZ`路由匹配器, 而不是自制匹配器, 以提升效率
- CHANGE: 使用`bodystream`进行req方向的body复制, 而不是使用额外的`buffer reader`
- CHANGE: 使用`HertZ`的`requestContext`传递matcher参数, 而不是`25w30a`中的ctx
- CHANGE: 改进`rate`模块, 避免并发竞争问题
- CHANGE: 将大部分状态码返回改为新的`html/tmpl`方式处理
- CHANGE: 修改部分log等级
- FIX:    修正默认配置的填充错误
- CHANGE: 使用go `html/tmpl`处理状态码页面, 同时实现错误信息显示
- CHANGE: 改进handle, 复用共同部分
- CHANGE: 细化url匹配的返回码处理
- CHANGE: 增加404界面

25w30e - 2025-04-24
---
- PRE-RELEASE: 此版本是v3.1.0预发布版本,请勿在生产环境中使用;
- CHANGE: 改进`rate`模块, 避免并发竞争问题
- CHANGE: 将大部分状态码返回改为新的`html/tmpl`方式处理
- CHANGE: 修改部分log等级
- FIX: 修正默认配置的填充错误

25w30d - 2025-04-22
---
- PRE-RELEASE: 此版本是v3.1.0预发布版本,请勿在生产环境中使用;
- CHANGE: 使用go `html/tmpl`处理状态码页面, 同时实现错误信息显示

25w30c - 2025-04-21
---
- PRE-RELEASE: 此版本是v3.1.0预发布版本,请勿在生产环境中使用;
- CHANGE: 改进handle, 复用共同部分
- CHANGE: 细化url匹配的返回码处理
- CHANGE: 增加404界面

25w30b - 2025-04-21
---
- PRE-RELEASE: 此版本是v3.1.0预发布版本,请勿在生产环境中使用;
- CHANGE: 使用`bodystream`进行req方向的body复制, 而不是使用额外的`buffer reader`
- CHANGE: 使用`HertZ`的`requestContext`传递matcher参数, 而不是`25w30a`中的标准ctx

25w30a - 2025-04-19
---
- PRE-RELEASE: 此版本是v3.1.0预发布版本,请勿在生产环境中使用;
- CHANGE: 对标准url使用`HertZ`路由匹配器, 而不是自制匹配器

3.0.3 - 2025-04-19
---
- CHANGE: 增加移除部分header的处置, 避免向服务端/客户端透露过多信息
- FIX: 修正非预期的header操作行为
- CHANGE: 合并header相关逻辑, 避免多次操作
- CHANGE: 对editor模式下的input进行处置, 增加隐式关闭处理
- CHANGE: 增加`netlib`配置项

25w29b - 2025-04-19
---
- PRE-RELEASE: 此版本是v3.0.3预发布版本,请勿在生产环境中使用;
- CHANGE: 增加`netlib`配置项

25w29a - 2025-04-17
---
- PRE-RELEASE: 此版本是v3.0.3预发布版本,请勿在生产环境中使用;
- CHANGE: 增加移除部分header的处置, 避免向服务端/客户端透露过多信息
- FIX: 修正非预期的header操作行为
- CHANGE: 合并header相关逻辑, 避免多次操作
- CHANGE: 对editor模式下的input进行处置, 增加隐式关闭处理

3.0.2 - 2025-04-15
---
- CHANGE: 避免重复的re编译操作
- CHANGE: 去除不必要的请求
- CHANGE: 改进`httpc`相关配置
- CHANGE: 更新`httpc` 0.4.0
- CHANGE: 为不遵守`RFC 2616`, `RFC 9112`的客户端带来兼容性改进

25w28b - 2025-04-15
---
- PRE-RELEASE: 此版本是v3.0.2预发布版本,请勿在生产环境中使用;
- CHANGE: 改进resp关闭
- CHANGE: 避免重复的re编译操作

25w28a - 2025-04-14
---
- PRE-RELEASE: 此版本是预发布版本,请勿在生产环境中使用;
- CHANGE: 去除不必要的请求
- CHANGE: 改进`httpc`相关配置
- CHANGE: 合入test版本修改

25w28t-2 - 2025-04-11
---
- TEST: 测试验证版本
- CHANGE: 为不遵守`RFC 2616`, `RFC 9112`的客户端带来兼容性改进

25w28t-1 - 2025-04-11
---
- TEST: 测试验证版本
- CHANGE: 更新httpc 0.4.0

3.0.1 - 2025-04-08
---
- CHANGE: 加入`memLimit`指示gc
- CHANGE: 加入`hlog`输出路径配置
- CHANGE: 修正H2C配置问题

25w27a - 2025-04-07
---
- PRE-RELEASE: 此版本是v3.0.1的预发布版本,请勿在生产环境中使用;
- CHANGE: 加入`memLimit`指示gc
- CHANGE: 加入`hlog`输出路径配置
- CHANGE: 修正H2C配置问题

3.0.0 - 2025-04-04
---
- RELEASE: Next Gen; 下一个起点; 
- CHANGE: 使用HertZ框架重构, 提升性能
- CHANGE: 前端在构建时加入, 新增`Design`,`Metro`,`Classic`主题
- CHANGE: 加入`Mino`主题对接选项
- FIX: 修正部分日志输出问题
- CHANGE: 移除gin残留
- CHANGE: 移除无用传入参数, 调整代码结构
- CHANGE: 改进cli
- CHANGE: 改进`脚本嵌套加速处理器`
- CHANGE&FIX: 使用`c.SetBodyStream`方式, 修正此前`chunked`传输中存在的诸多问题, 参看[HertZ Issues #1309](https://github.com/cloudwego/hertz/issues/1309)
- PORT: 从v2移植`matcher`相关改进
- CHANGE: 增加默认配置生成
- CHANGE: 优化前端资源加载
- CHANGE: 将`cfg`flag改为`c`以符合`POSIX`规范
- CHANGE: 为`smart-git`添加`no-cache`标头

25w26a - 2025-04-03
---
- PRE-RELEASE: 此版本是v3的预发布版本,请勿在生产环境中使用;

2.6.3 - 2025-03-30
---
- FIX: 修正一些`git clone`行为异常

25w25a - 2025-03-30
---
- PRE-RELEASE: 此版本是v2.6.3的预发布版本,请勿在生产环境中使用;
- FIX: 修正一些`git clone`行为异常

e3.0.7 -2025-03-29
---
- CHANGE: 将`cfg`flag改为`c`以符合`POSIX`规范
- CHANGE: 为`smart-git`添加`no-cache`标头

2.6.2 - 2025-03-29
---
- BACKPORT: 反向移植前端资源加载改进

e3.0.6 - 2025-03-28
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- FIX: 修正状态码相关问题(开发遗留所致)

e3.0.5 - 2025-03-28
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- CHANGE: 增加默认配置生成
- CHANGE: 优化前端资源加载

2.6.1 - 2025-03-27
---
- CHANGE: 改进`matcher`组件
- CHANGE: 加入优雅关闭

e3.0.3 - 2025-03.27
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- E-RELEASE: 修正过往问题, 还请各位多多测试反馈
- PORT: 从v2移植`matcher`相关改进
- CHANGE&FIX: 使用`c.SetBodyStream`方式, 修正此前`chunked`传输中存在的诸多问题, 参看[HertZ Issues #1309](https://github.com/cloudwego/hertz/issues/1309)

25w24a - 2025-03-27
---
- PRE-RELEASE: 此版本是v2.6.1的预发布版本,请勿在生产环境中使用;
- CHANGE: 改进`matcher`组件
- CHANGE: 加入优雅关闭

e3.0.3rc2 - 2025-03-27
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- PRE-RELEASE: 此版本是v3.0.3的候选版本,请勿在生产环境中使用;
- PORT: 从v2移植`matcher`相关改进

e3.0.3rc1 - 2025-03-26
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- PRE-RELEASE: 此版本是v3.0.3的候选版本,请勿在生产环境中使用;
- CHANGE&FIX: 使用`c.SetBodyStream`方式, 修正此前`chunked`传输中存在的诸多问题, 参看[HertZ Issues #1309](https://github.com/cloudwego/hertz/issues/1309)

2.6.0 - 2025-03-22
---
- BACKPORT: 将v3的功能性改进反向移植

25w23a - 2025-03-22
---
- PRE-RELEASE: 此版本是v2.6.0的预发布版本,请勿在生产环境中使用;
- BACKPORT: 将v3的功能性改进反向移植

e3.0.2 - 2025-03-21
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- RELEASE: 在此表达对各位的歉意, v3迁移到HertZ带来了许多问题; 此版本完善v3的同时, 修正已知问题;
- FIX: 使用等效`c.Writer()`, 回归v2.5.0 func以修正问题
- CHANGE: 更新相关依赖

25w22a - 2025-03-21
---
- PRE-RELEASE: 此版本是v3.0.1的预发布版本,请勿在生产环境中使用;
- FIX: 使用等效`c.Writer()`, 回归v2.5.0 func以修正问题

e3.0.1 - 2025-03-21
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- RELEASE: Next Step; 下一步; 完善v3的同时, 修正已知问题;
- CHANGE: 改进cli
- CHANGE: 重写`ProcessLinksAndWriteChunked`(脚本嵌套加速处理器), 修正已知问题的同时提高性能与效率
- CHANGE: 完善`gitreq`部分
- FIX: 修正日志输出格式问题
- FIX: 使用更新的`hwriter`以修正相关问题

25w21e - 2025-03-21
---
- PRE-RELEASE: 此版本是v3.0.1的预发布版本,请勿在生产环境中使用;
- CHANGE: 重写`ProcessLinksAndWriteChunked`(脚本嵌套加速处理器), 修正已知问题的同时提高性能与效率

25w21d - 2025-03-21
---
- PRE-RELEASE: 此版本是v3.0.1的预发布版本,请勿在生产环境中使用;
- FIX: 使用更新的`hwriter`以修正相关问题

25w21c - 2025-03-20
---
- PRE-RELEASE: 此版本是v3.0.1的预发布版本,请勿在生产环境中使用;
- TEST: 测试新的`hwriter`

25w21b - 2025-03-20
---
- PRE-RELEASE: 此版本是v3.0.1的预发布版本,请勿在生产环境中使用;
- FIX: 修正日志输出格式问题

25w21a - 2025-03-20
---
- PRE-RELEASE: 此版本是v3.0.1的预发布版本,请勿在生产环境中使用;
- CHANGE: 改进cli
- CHANGE: 完善`gitreq`部分

e3.0.0 - 2025-03-19
---
- ATTENTION: 此版本是实验性的, 请确保了解这一点
- RELEASE: Next Gen; 下一个起点; 
- CHANGE: 使用HertZ框架重构, 提升性能
- CHANGE: 前端在构建时加入, 新增`Design`,`Metro`,`Classic`主题
- CHANGE: 加入`Mino`主题对接选项
- FIX: 修正部分日志输出问题
- CHANGE: 移除gin残留
- CHANGE: 移除无用传入参数, 调整代码结构

25w20b - 2025-03-19
---
- PRE-RELEASE: 此版本是v3.0.0的预发布版本,请勿在生产环境中使用; 
- CHANGE: 加入`Mino`主题对接选项
- FIX: 修正部分日志输出问题
- CHANGE: 移除gin残留
- CHANGE: 移除无用传入参数, 调整代码结构

25w20a - 2025-03-18
---
- PRE-RELEASE: 此版本是v3.0.0的预发布版本,请勿在生产环境中使用; 
- CHANGE: 使用HertZ重构
- CHANGE: 前端在构建时加入, 新增`Design`,`Metro`,`Classic`主题

2.5.0 - 2025-03-17
---
- ADD: 加入脚本嵌套加速功能
- CHANGE: 改进Auth模块

25w19a - 2025-03-16
---
- PRE-RELEASE: 此版本是v2.5.0的预发布版本,请勿在生产环境中使用;
- ADD: 加入脚本嵌套加速功能
- CHANGE: 改进Auth模块
- CHANGE: 将handler模块化改进

2.4.2 - 2025-03-14
---
- CHANGE: 在GitClone Cache模式下, 相关请求会使用独立httpc client
- CHANGE: 为GitClone Cache的独立httpc client增加ForceH2C选项
- FIX: 修正GitClone Cache模式下的Url生成问题

25w18a - 2025-03-14
---
- PRE-RELEASE: 此版本是v2.4.2的预发布版本,请勿在生产环境中使用;
- CHANGE: 在GitClone Cache模式下, 相关请求会使用独立httpc client
- CHANGE: 为GitClone Cache的独立httpc client增加ForceH2C选项
- FIX: 修正GitClone Cache模式下的Url生成问题

2.4.1 - 2025-03-13
---
- CHANGE: 重构路由匹配
- CHANGE: 更新相关依赖以修复错误

25w17a - 2025-03-13
---
- PRE-RELEASE: 此版本是v2.4.1的预发布版本,请勿在生产环境中使用;
- CHANGE: 重构路由匹配
- CHANGE: 更新相关依赖以修复错误

2.4.0 - 2025-03-12
---
- ADD: 支持通过[Smart-Git](https://github.com/WJQSERVER-STUDIO/smart-git)实现Git Clone缓存
- CHANGE: 使用更高性能的Buffer Pool 实现, 调用 github.com/WJQSERVER-STUDIO/go-utils/copyb
- CHANGE: 改进路由匹配
- CHANGE: 更新依赖
- CHANGE: 改进前端

25w16d - 2025-03-12
---
- PRE-RELEASE: 此版本是v2.4.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 使用更高性能的Buffer Pool 实现

25w16c
---
- PRE-RELEASE: 此版本是v2.4.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 使用更高性能的Buffer Pool 实现
- CHANGE: 改进路由匹配

25w16b
---
- PRE-RELEASE: 此版本是v2.4.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 修改路由
- CHANGE: 改进前端

25w16a
---
- PRE-RELEASE: 此版本是v2.4.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 变更CORS配置
- ADD: 使用GO-GIT实现git smart http服务端和客户端
- CHANGE: 更新依赖

2.3.1
---
- CHANGE: 改进`Pages`在`External`模式下的路由
- CHANGE: 使用`H2C` bool 代替 `enableH2C` string (2.4.0 弃用 `enableH2C`)
- CHANGE: 使用`Mode` string 代替`Pages`内的 `enable` bool (2.4.0 弃用 `enable`)

25w15a
---
- PRE-RELEASE: 此版本是v2.3.1的预发布版本,请勿在生产环境中使用;
- CHANGE: 改进`Pages`在`External`模式下的路由
- CHANGE: 使用`H2C` bool 代替 `enableH2C` string (2.4.0 弃用 `enableH2C`)
- CHANGE: 使用`Mode` string 代替`Pages`内的 `enable` bool (2.4.0 弃用 `enable`)

2.3.0
---
- CHANGE: 使用`touka-httpc`封装`HTTP Client`, 更新到`v0.2.0`版本, 参看`touka-httpc`
- CHANGE: 重构前端页面, 见[#49](https://github.com/WJQSERVER-STUDIO/ghproxy/pull/49)
- CHANGE: 重构`blacklist`实现
- CHANGE: 优化404处理
- CHANGE: 重构`whitelist`实现
- CHANGE: 对`proxy`进行结构性调整
- CHANGE: `chunckedreq`与`gitreq`共用`BufferPool`和`HTTP Client`
- CHANGE: 新增`HTTP Client`配置块
- CHANGE: 加入内置主题配置, 支持通过配置切换主题
- CHANGE: 将许可证转为WJQserver Studio License 2.0

25w14b
---
- PRE-RELEASE: 此版本是v2.3.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 将许可证转为WJQserver Studio License 2.0

25w14a
---
- PRE-RELEASE: 此版本是v2.3.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 使用`touka-httpc`封装`HTTP Client`, 更新到`v0.2.0`版本, 参看`touka-httpc`
- CHANGE: 重构前端页面, 见[#49](https://github.com/WJQSERVER-STUDIO/ghproxy/pull/49)
- CHANGE: 重构`blacklist`实现
- CHANGE: 优化404处理
- CHANGE: 重构`whitelist`实现
- CHANGE: 对`proxy`进行结构性调整
- CHANGE: `chunckedreq`与`gitreq`共用`BufferPool`和`HTTP Client`
- CHANGE: 新增`HTTP Client`配置块
- CHANGE: 加入内置主题配置, 支持通过配置切换主题

25w14t-2
---
- PRE-RELEASE: 此版本是测试验证版本,请勿在生产环境中使用;
- CHANGE: 使用`touka-httpc`封装`HTTP Client`，更新到`v0.1.0`版本, 参看`touka-httpc`
- CHANGE: 重构`whitelist`实现
- CHANGE: 对`proxy`进行结构性调整
- CHANGE: `chunckedreq`与`gitreq`共用`BufferPool`和`HTTP Client`
- CHANGE: 新增`HTTP Client`配置块

25w14t-1
---
- PRE-RELEASE: 此版本是测试验证版本,请勿在生产环境中使用;
- CHANGE: 使用`touka-httpc`封装`HTTP Client`
- CHANGE: 重构前端页面, 见[#49](https://github.com/WJQSERVER-STUDIO/ghproxy/pull/49)
- CHANGE: 重构`blacklist`实现
- CHANGE: 优化404处理

2.2.0
---
- RELEASE: v2.2.0正式版发布;
- CHANGE: 更新Go版本至1.24.0
- ADD: 加入`Socks5`和`HTTP(S)`出站支持
- CHANGE: 配置新增`Outbound`配置块

25w13b
---
- PRE-RELEASE: 此版本是v2.2.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 更新Go版本至1.24.0

25w13a
---
- PRE-RELEASE: 此版本是v2.2.0的预发布版本,请勿在生产环境中使用;
- ADD: 加入`Socks5`和`HTTP(S)`出站支持
- CHANGE: 配置新增`Outbound`配置块

2.1.0
---
- RELEASE: v2.1.0正式版发布;
- CHANGE: 加入`FreeBSD`与`Darwin`系统支持
- CHANGE: 更新安全政策, v1和24w版本序列生命周期正式结束
- ADD: 加入`timing`中间件记录响应时间
- ADD: 加入`loggin`中间件包装日志输出
- CHANGE: 更新logger版本至v1.3.0
- CHANGE: 改进日志相关
- ADD: 加入日志等级配置项

25w12d
---
- PRE-RELEASE: 此版本是v2.1.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 处理类型断言相关问题

25w12c
---
- PRE-RELEASE: 此版本是v2.1.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 加入`FreeBSD`与`Darwin`系统支持

25w12b
---
- PRE-RELEASE: 此版本是v2.0.8/v2.1.0的预发布版本,请勿在生产环境中使用;
- ADD: 加入`timing`中间件记录响应时间
- ADD: 加入`loggin`中间件包装日志输出
- CHANGE: 更新安全政策, v1和24w版本序列生命周期正式结束

25w12a
---
- PRE-RELEASE: 此版本是v2.0.8/v2.1.0的预发布版本,请勿在生产环境中使用;
- CHANGE: 更新logger版本至v1.3.0
- CHANGE: 改进日志相关
- ADD: 加入日志等级配置项

2.0.7
---
- RELEASE: v2.0.7正式版发布;
- CHANGE: 更新Go版本至1.23.6
- CHANGE: 更新Logger版本至v1.2.0

25w11a
---
- PRE-RELEASE: 此版本是v2.0.7的预发布版本,请勿在生产环境中使用;
- CHANGE: 更新Go版本至1.23.6
- CHANGE: 更新Logger版本至v1.2.0

2.0.6
---
- RELEASE: v2.0.6正式版发布;祝各位新春快乐!
- CHANGE: 优化前端的连接转换逻辑
- CHANGE: 优化代码内不必要的函数化, 1.4之后, 函数化疑似有点太多了
- 优化`HTTP Client`参数
- CHANGE: 为api路由组增加no-cache标头

25w10b
---
- PRE-RELEASE: 此版本是v2.0.6的预发布版本,请勿在生产环境中使用;祝各位新春快乐!
- CHANGE: 为api路由组增加no-cache标头

25w10a
---
- PRE-RELEASE: 此版本是v2.0.6的预发布版本,请勿在生产环境中使用;祝各位新春快乐!
- CHANGE: 改进前端的连接转换逻辑
- CHANGE: 优化代码内不必要的函数化, 1.4之后, 函数化疑似有点太多了
- 优化`HTTP Client`参数

2.0.5
---
- RELEASE: v2.0.5正式版发布;
- CHANGE: 优化响应体分块复制实现
- ADD: 加入缓存池
- CHANGE: 改进缓存实现
- CHANGE: 部分杂项改进

25w09b
---
- PRE-RELEASE: 此版本是v2.0.5的预发布版本,请勿在生产环境中使用;
- REMOVE: 移除残留配置

25w09a
---
- PRE-RELEASE: 此版本是v2.0.5的预发布版本,请勿在生产环境中使用;
- CHANGE: 改进缓存实现
- ADD: 加入缓存池

2.0.4
---
- RELEASE: v2.0.4正式版发布;
- CHANGE: 优化GitReq的`HTTP Client`参数, 使其更符合本项目使用场景
- CHANGE: 优化Matches
- REMOVE: 移除Caddyfile残留
- REMOVE: 由于v2改进后稳定性增强, 故移除健康检测

25w08b
---
- PRE-RELEASE: 此版本是v2.0.4的预发布版本,请勿在生产环境中使用;
- REMOVE: 由于v2改进后稳定性增强, 故移除健康检测

25w08a
---
- PRE-RELEASE: 此版本是v2.0.4的预发布版本,请勿在生产环境中使用;
- CHANGE: 优化GitReq的`HTTP Client`参数, 使其更符合本项目使用场景
- CHANGE: 优化Matches
- REMOVE: 移除Caddyfile残留

2.0.3
---
- RELEASE: v2.0.3正式版发布;
- CHANGE: 优化`HTTP Client`参数, 使其更符合本项目使用场景

25w07b
---
- PRE-RELEASE: 此版本是v2.0.3的预发布版本,请勿在生产环境中使用;
- CHANGE: 改进`HTTP Client`参数

25w07a
---
- PRE-RELEASE: 此版本是v2.0.3的预发布版本,请勿在生产环境中使用;
- CHANGE: 为`HTTP Client`增加复用, 对性能有所优化
- CHANGE: 优化`HTTP Client`参数, 使其更符合本项目使用场景

2.0.2
---
- RELEASE: v2.0.2正式版发布; 此版本是v2.0.1改进
- CHANGE: 由于用户使用了不符合`RFC 9113`规范的请求头, 导致`ghproxy`无法正常工作, 在此版本为用户的错误行为提供补丁; 

25w06b
---
- PRE-RELEASE: 此版本是改进验证版本,普通用户请勿使用; 
- CHANGE: 由于用户使用了不符合`RFC 9113`规范的请求头, 导致`ghproxy`无法正常工作, 在此版本为用户的错误行为提供补丁; 

25w06a
---
- PRE-RELEASE: 此版本是改进验证版本,普通用户请勿使用; 
- CHANGE: Remove `Conection: Upgrade` header, which is not currently supported by some web server configurations.

v2.0.1
---
- RELEASE: v2.0.1正式版发布; 此版本是v2.0.0的小修复版本, 主要修复了Docker启动脚本存在的一些问题
- FIX: 修复Docker启动脚本存在的一些问题

25w05a
---
- PRE-RELEASE: 此版本是v2.0.1的候选版本,请勿在生产环境中使用;
- FIX: 修复Docker启动脚本存在的一些问题

2.0.0
---
- RELEASE: v2.0.0正式版发布; 此版本圆了几个月前画的饼, 在大文件下载的内存占用方面做出了巨大改进
- CHANGE: 优化`proxy`核心模块, 使用Chuncked Buffer传输数据, 减少内存占用
- REMOVE: caddy
- REMOVE: nocache
- CHANGE: 优化前端页面, 增加更多功能(来自1.8.1版本, 原本也是为v2所设计的)

25w04c
---
- PRE-RELEASE: 此版本是v2的候选版本,请勿在生产环境中使用;
- CHANGE: 大幅优化`proxy`核心模块, 使用Chuncked Buffer传输数据, 减少内存占用

v1.8.3
---
- RELEASE: v1.8.3, 此版本作为v1.8.2的依赖更新版本(在v2发布前, v1仍会进行漏洞修复)
- CHANGE: 更新Go版本至`1.23.5`以解决CVE漏洞

25w04b
---
- PRE-RELEASE: 此版本是v2的候选版本(技术验证版),请勿在生产环境中使用; 我们可能会撤除v2更新计划(若技术验证版顺利通过, 则会发布v2正式版)
- REMOVE: caddy

25w04a
---
- PRE-RELEASE: 此版本是v2的候选版本(技术验证版),请勿在生产环境中使用; 我们可能会撤除v2更新计划(若技术验证版顺利通过, 则会发布v2正式版)
- CHANGE: 大幅修改核心组件

1.8.2
---
- RELEASE: v1.8.2正式版发布; 这或许会是v1的最后一个版本
- FIX: 修复部分日志表述错误
- CHANGE: 关闭`gin`框架的`fmt`日志打印, 在高并发场景下提升一定性能(go 打印终端日志性能较差，可能造成性能瓶颈)

25w03a
---
- PRE-RELEASE: 此版本是v1.8.2的候选预发布版本,请勿在生产环境中使用
- FIX: 修复部分日志表述错误
- CHANGE: 关闭`gin`框架的`fmt`日志打印, 在高并发场景下提升一定性能(go 打印终端日志性能较差，可能造成性能瓶颈)

1.8.1
---
- RELEASE: v1.8.1正式版发布; 此版本发布较为仓促, 用于修复caddy2.9.0导致的问题
- CHANGE: 更新底包至`v2.9.1`
- FIX: 修复caddy2.9.0导致的问题
- CHANGE: 对前端进行重构(非最终决定,各位可将其与原先的版本对比, 若有相关建议, 请与开发团队进行交流)

25w02a
---
- PRE-RELEASE: 此版本是v1.8.1的候选预发布版本,请勿在生产环境中使用
- CHANGE: 更新底包至`v2.9.1`
- CHANGE: 对前端进行重构(非最终决定,各位可将其与原先的版本对比, 若有相关建议, 请与开发团队进行交流)

v1.8.0
---
- RELEASE: v1.8.0正式版发布; 这是2025年的第一个正式版本发版,祝各位新年快乐!
- CHANGE: 更新底包至`v2.9.0`
- CHANGE: 优化`Auth`参数透传至`"Authorization: token {token}"`功能, 增加`dev`参数以便调试
- CHANGE: 优化`config.toml`默认配置, 增加`embed.FS`内嵌前端支持, 并优化相关逻辑
- CHANGE: 更新前端页面版权声明

25w01e
---
- PRE-RELEASE: 此版本是v1.8.0的预发布版本,请勿在生产环境中使用
- FIX: 修复引入token参数透传功能导致的一些问题

25w01d
---
- PRE-RELEASE: 此版本是v1.8.0的预发布版本,请勿在生产环境中使用
- CHANGE: 尝试修复部分问题

25w01c
---
- PRE-RELEASE: 此版本是v1.8.0的预发布版本,请勿在生产环境中使用
- CHANGE: 改进token参数透传功能

25w01b
---
- PRE-RELEASE: 此版本是v1.8.0的预发布版本,请勿在生产环境中使用
- CHANGE: 将底包更新至`v2.9.0`

25w01a
---
- PRE-RELEASE: 此版本是v1.8.0的预发布版本,请勿在生产环境中使用; 同时,这也是2025年的第一个pre-release版本,祝各位新年快乐! (同时,请注意版本号的变化)
- ADD: 加入`dev`参数, 以便pre-release版本调试(实验性)
- ADD: 加入基于`embed.FS`的内嵌前端, config.toml中的`[pages]`配置为false时自动使用内嵌前端
- CHANGE: 完善24w29a版本新加入的`Auth`参数透传至`"Authorization: token {token}"`功能，对相关逻辑进行完善
- FIX: 修正24w29a版本新加入的`Auth`参数透传至`"Authorization: token {token}"`功能的一些问题

24w29a
---
- PRE-RELEASE: 此版本是一个实验性功能测试版本,请勿在生产环境中使用; 同时,这也是2024年的最后一个pre-release版本
- ADD: `Auth` token参数透传至`"Authorization: token {token}"`, 为私有仓库拉取提供一定便利性(需要更多测试)
- CHANGE: 更新相关依赖库

v1.7.9
---
- RELEASE: 安全性及小型修复, 建议用户自行选择是否升级
- CHANGE: 将`logger`库作为外部库引入, 使维护性更好, 同时修正了部分日志问题并提升部分性能
- CHANGE: 更新相关依赖库, 更新`req`库以解决`net`标准库的`CVE-2021-38561`漏洞
- FIX: 修复安装脚本内的错误

24w28b
---
- PRE-RELEASE: 此版本是v1.7.9的预发布版本,请勿在生产环境中使用
- CHANGE: 将`logger`库作为外部库引入, 使维护性更好, 同时修正了部分日志问题并提升部分性能

24w28a
---
- PRE-RELEASE: 此版本是v1.7.9的预发布版本,请勿在生产环境中使用
- CHANGE: 更新相关依赖库, 更新`req`库以解决`net`标准库的`CVE-2021-38561`漏洞
- FIX: 修复安装脚本内的错误

v1.7.8
---
- RELEASE: 我们建议您升级到此版本, 以解决一些依赖库的安全漏洞和与caddy相关的内存问题
- CHANGE: 更新底包至`v24.12.20`可能解决部分与`caddy`相关的内存问题
- CHANGE: 更新相关依赖库,解决`net`标准库的`CVE-2024-45338`
- CHANGE: 小幅更新前端页面
- FIX: 修复`config.toml`默认配置内的错误
- ADD: 新增`api.github.com`反代支持, 强制性要求开启`Header Auth`功能(需要更多测试)

24w27e
---
- PRE-RELEASE: 此版本是v1.7.8的预发布候选版本(若无问题,此版本将会成为v1.7.8正式版本),请勿在生产环境中使用
- CHANGE: 更新底包至`v24.12.20`可能解决部分与`caddy`相关的内存问题

24w27d
---
- PRE-RELEASE: 此版本是v1.7.8的预发布候选版本,请勿在生产环境中使用
- CHANGE: 更新相关依赖库,解决`net`标准库的`CVE-2024-45338`
- CHANGE: 小幅更新前端页面

24w27c
---
- PRE-RELEASE: 此版本做为实验性功能测试版本,请勿在生产环境中使用
- CHANGE: 更新docker底包至`v2.9.0-beta.3` , 可能解决部分内存相关问题
- CHANGE: 更新相关依赖库

24w27b
---
- PRE-RELEASE: 此版本做为实验性功能测试版本,请勿在生产环境中使用
- FIX: 修复`config.toml`默认配置内的错误

24w27a
---
- PRE-RELEASE: 此版本做为实验性功能测试版本,请勿在生产环境中使用
- ADD: 新增`api.github.com`反代支持, 强制性要求开启`Header Auth`功能

v1.7.7
---
- CHANGE: 更新相关依赖库
- CHANGE: 更新Go版本至1.23.4
- CHANGE: 更新release及dev版本底包

24w26a
---
- PRE-RELEASE: 此版本是v1.7.7的预发布版本,请勿在生产环境中使用
- CHANGE: 更新相关依赖库
- CHANGE: 更新Go版本至1.23.4
- CHANGE: 更新release及dev版本底包

v1.7.6
---
- RELEASE: 版本在v1.7.4及以上的用户,我们建议升级到此版本以解决于v1.7.4版本功能更新所引入的问题
- FIX: 进一步修正 H2C相关配置逻辑问题
- CHANGE: 对Caddy配置进行实验性修改,优化H2C配置
- CHANGE: 更新相关依赖库

24w25b
---
- PRE-RELEASE: 此版本是v1.7.6的预发布版本,请勿在生产环境中使用
- 说明: 本版本为24w25a-fix0
- FIX: 进一步修正 H2C相关配置逻辑问题

24w25a
---
- PRE-RELEASE: 此版本是v1.7.6的预发布版本,请勿在生产环境中使用
- 说明: 本版本为v1.7.6的其中一个候选与开发测试版本,相关改动不一定实装
- FIX: 进一步修正 H2C相关配置逻辑问题
- CHANGE: 对Caddy配置进行实验性修改,优化H2C配置
- CHANGE: 更新相关依赖库

v1.7.5
---
- FIX: 修复 v1.7.4 版本 Docker 镜像默认配置导致的 403 问题
- ADD: `Rate`模块加入`IP`速率限制,可限制单个IP的请求速率 (需要更多测试)
- CHANGE: 处理积攒的依赖库更新

24w24c
---
- PRE-RELEASE: 此版本是v1.7.5的预发布版本,请勿在生产环境中使用
- CHANGE: 更新依赖

24w24b
---
- PRE-RELEASE: 此版本是v1.7.5的预发布版本,请勿在生产环境中使用
- FIX: 修复 Docker 默认配置导致的 403 问题

24w24a
---
- PRE-RELEASE: 此版本是v1.7.5的预发布版本,请勿在生产环境中使用
- ADD: `Rate`模块加入`IP`速率限制,可限制单个IP的请求速率 (需要更多测试)
- CHANGE: 处理积攒的依赖库更新,更新如下依赖库:
- **github.com/gabriel-vasile/mimetype**: 从 v1.4.6 升级到 v1.4.7
- **github.com/go-playground/validator/v10**: 从 v10.22.1 升级到 v10.23.0
- **github.com/klauspost/cpuid/v2**: 从 v2.2.8 升级到 v2.2.9
- **github.com/onsi/ginkgo/v2**: 从 v2.21.0 升级到 v2.22.0
- **golang.org/x/arch**: 从 v0.11.0 升级到 v0.12.0
- **golang.org/x/crypto**: 从 v0.28.0 升级到 v0.29.0
- **golang.org/x/exp**: 从 v0.0.0-20241009180824-f66d83c29e7c 升级到 v0.0.0-20241108190413-2d47ceb2692f
- **golang.org/x/mod**: 从 v0.21.0 升级到 v0.22.0
- **golang.org/x/net**: 从 v0.30.0 升级到 v0.31.0
- **golang.org/x/sync**: 从 v0.8.0 升级到 v0.9.0
- **golang.org/x/sys**: 从 v0.26.0 升级到 v0.27.0
- **golang.org/x/text**: 从 v0.19.0 升级到 v0.20.0
- **golang.org/x/tools**: 从 v0.26.0 升级到 v0.27.0
- **google.golang.org/protobuf**: 从 v1.35.1 升级到 v1.35.2

v1.7.4
---
- CHANGE: 对二进制文件部署脚本进行优化
- CHANGE&ADD: 新增H2C相关配置
- ADD: `Auth`模块加入`Header`鉴权,使用`GH-Auth`的值进行鉴权

24w23a
---
- PRE-RELEASE: 此版本是v1.7.4的预发布版本,请勿在生产环境中使用
- ADD: `Auth`模块加入`Header`鉴权,使用`GH-Auth`的值进行鉴权
- CHANGE: 对二进制文件部署脚本进行优化
- CHANGE&ADD: 新增H2C相关配置

v1.7.3
---
- CHANGE: Bump golang.org/x/time from 0.7.0 to 0.8.0
- FIX: 修复故障熔断的相关问题

v1.7.2
---
- CHANGE: 为`nocache`版本加入测试性的故障熔断机制

v1.7.1
---
- CHANGE: 更新Go版本至1.23.3
- CHANGE: 更新相关依赖库
- ADD: 对`Proxy`模块进行优化,增加使用`HEAD`方式预获取`Content-Length`头
- CHANGE: 将`release`与`dev`版本的底包切换至`wjqserver/caddy:2.9.0-rc4-alpine`，将`nocache`版本的底包切换至`alpine:latest`
- CHANGE: 对`nocache`版本的`config.toml`与`init.sh`进行适配性修改
- CHANGE: 加入测试性的故障熔断机制(Failure Circuit Breaker) (nocache版本暂不支持)

24w22b
---
- PRE-RELEASE: 此版本是v1.7.1的预发布版本,请勿在生产环境中使用
- CHANGE: 更新Go版本至1.23.3
- CHANGE: 更新相关依赖库
- ADD: 对`Proxy`模块进行优化,增加使用`HEAD`方式预获取`Content-Length`头
- CHANGE: 将`release`与`dev`版本的底包切换至`wjqserver/caddy:2.9.0-rc4-alpine`，将`nocache`版本的底包切换至`alpine:latest`
- CHANGE: 对`nocache`版本的`config.toml`与`init.sh`进行适配性修改

24w22a
---
- PRE-RELEASE: 此版本是v1.7.1的预发布版本,请勿在生产环境中使用
- CHANGE: 更新底包
- CHANGE: 加入测试性的故障熔断机制(Failure Circuit Breaker)

v1.7.0
---
- ADD: 加入`rate`模块,实现内置速率限制
- CHANGE: 优化`blacklist`与`whitelist`模块的匹配算法,提升性能；由原先的完整匹配改为切片匹配，提升匹配效率
- ADD: 加入`version`相关表示与API接口
- ADD: 加入`rate`相关API接口
- CHANGE: 优化前端界面,优化部分样式
- CHANGE: 更新相关依赖库
- CHANGE: 对编译打包进行改进,此后不再提供独立可执行文件,请改为拉取`tar.gz`压缩包

24w21d
---
- PRE-RELEASE: 此版本是v1.7.0的预发布版本,请勿在生产环境中使用
- ADD: 新增`ratePerMinute` API可供查询
- ADD: 前端新增 version 标识
- ADD: 前端新增 `重定向` 按钮,用于重定向到代理后的链接
- CHANGE: 优化输出代码块,使样式更加美观
- CHANGE: 更新相关依赖库
- CHANGE: 对黑名单模块进行实验性功能优化,提升性能(改进匹配算法,在切片后优先匹配user,减少无效匹配)

24w21c
---
- PRE-RELEASE: 此版本是v1.7.0的预发布版本,请勿在生产环境中使用
- CHANGE: 对编译打包进行改进,此后不再提供独立可执行文件,请改为拉取`tar.gz`压缩包
- CHANGE: 由于上述原因,对Docker打包进行相应改进

24w21b
---
- PRE-RELEASE: 此版本是v1.7.0的预发布版本,请勿在生产环境中使用
- ADD: 加入版本号标识与对应API接口
- ADD: 加入速率限制API接口
- CHANGE: 修改打包部分

24w21a
---
- PRE-RELEASE: 此版本是v1.7.0的预发布版本,请勿在生产环境中使用
- ADD: 尝试加入程序内置速率限制
- CHANGE: 更新相关依赖库
- CHANGE: 更换Dev版本底包,于release版本保持一致

v1.6.2
---
- CHANGE: 优化前端界面,优化部分样式
- ADD: 前端加入黑夜模式
- CHANGE: 优化移动端适配
- CHANGE: 优化一键部署脚本,使其更加易用,并增加更多的功能(已于早些时候hotfix)
- CHANGE: 优化部分代码结构,提升性能
- CHANGE: 优化日志记录,对各个部分的日志记录进行统一格式,并对部分重复日志进行合并

24w20b
---
- PRE-RELEASE: 此版本是v1.6.2的预发布版本,请勿在生产环境中使用
- CHANGE: 优化前端界面,优化部分样式
- ADD: 前端加入黑夜模式
- CHANGE: 优化移动端适配

24w20a
---
- PRE-RELEASE: 此版本是v1.6.2的预发布版本,请勿在生产环境中使用
- CHANGE: 大幅修改日志记录,对各个部分的日志记录进行统一格式,并对部分重复日志进行合并
- CHANGE: 大幅优化一键部署脚本,使其更加易用,并增加更多的功能(已于早些时候hotfix)
- CHANGE: 优化部分代码结构,提升性能

v1.6.1
---
- CHANGE: 根据社区建议,将`sizeLimit`由过去的以`byte`为单位,改为以`MB`为单位,以便于直观理解
- ADD: 新增`nocache`版本,供由用户自行优化缓存策略
- CHANGE: 优化`Proxy`核心模块内部结构,提升性能
- REMOVE: 移除`Proxy`模块内部分无用`logInfo`
- FIX & ADD: 修复前端对gist的匹配问题,添加对`gist.githubusercontent.com`的前端转换支持
- CHANGE: 改变部分前端匹配逻辑
- CHANGE: 更新相关依赖库

24w19d
---
- PRE-RELEASE: 此版本是v1.6.1的预发布版本,请勿在生产环境中使用
- ADD: 新增nocache版本,供由用户自行优化缓存策略
- CHANGE: 优化`Proxy`核心模块内部结构,提升性能
- REMOVE: 移除`Proxy`模块内部分无用`logInfo`

24w19c
---
- PRE-RELEASE: 此版本是v1.6.1的预发布版本,请勿在生产环境中使用
- FIX & ADD: 修复前端对gist的匹配问题,添加对`gist.githubusercontent.com`的前端转换支持
- CHANGE: 改变部分前端匹配逻辑
- CHANGE: 更新相关依赖库

24w19b
---
- PRE-RELEASE: 此版本是v1.6.1的预发布版本,请勿在生产环境中使用
- FIX: 修复`sizeLimit`单位更改导致API返回值错误的问题
- FIX: 修正Gist匹配

24w19a
---
- PRE-RELEASE: 此版本是v1.6.1的预发布版本,请勿在生产环境中使用
- CHANGE: 根据社区建议,将`sizeLimit`由过去的以`byte`为单位,改为以`MB`为单位,以便于直观理解
- CHANGE: 更新相关依赖
- CHANGE: 对`Proxy`模块的核心函数进行模块化,为后续修改和扩展提供空间

v1.6.0
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: 引入H2C支持,支持无加密HTTP/2请求,一定程度上提升传输性能
- ADD: 在核心程序内加入静态页面支持,支持不通过caddy等web server提供前端页面
- CHANGE: 优化日志记录,带来更多的可观测性
- CHANGE: 改进前端界面,优化用户体验; 对原有Alert提示进行优化，改为ShowToast提示
- CHANGE: 规范化部分函数命名,提升可读性; 同时对config.toml内的参数命名进行规范化(部分参数名称已过时，请注意更新)
- CHANGE: 修改日志检查周期,降低检查频率,避免不必要的资源浪费
- ADD: 增加CORS状态API

24w18f
---
- PRE-RELEASE: 此版本是v1.6.0的预发布版本,请勿在生产环境中使用
- CHANGE: 修正前端页面的部分样式问题
- FIX: 修正部分问题

24w18e
---
- PRE-RELEASE: 此版本是预发布版本,请勿在生产环境中使用
- CHANGE: 引入H2C协议支持,支持无加密HTTP/2请求
- ADD: 尝试在核心程序内加入静态页面支持
- CHANGE: 优化日志记录
- CHANGE: 去除部分无用/重复配置
- CHANGE: 规范化部分函数命名

24w18d
---
- PRE-RELEASE: 此版本是预发布版本,请勿在生产环境中使用
- CHANGE: 更新相关依赖库
- ADD: 增加CORS状态API
- CHANGE: 优化部分函数执行顺序
- CHANGE: 优化前端界面

24w18c
---
- PRE-RELEASE: 此版本是预发布版本,请勿在生产环境中使用
- CHANGE: 修正配置命名,改为驼峰式命名
- CHANGE: 修正函数命名

24w18b
---
- PRE-RELEASE: 此版本是预发布版本,请勿在生产环境中使用
- CHANGE: 经团队考量,移除 Docker 代理功能，若造成了不便敬请谅解
- CHANGE: 修改日志检查周期

24w18a
---
- PRE-RELEASE: 此版本是预发布版本,请勿在生产环境中使用
- CHANGE: 改进Docker 代理
- CHANGE: 改进前端页面的copy提示,弃用alert提示

v1.5.2
---
- FIX: 修正flag传入问题
- CHANGE: 去除/路径重定向,改为返回403,并记录对应请求日志
- CHANGE: 优化Proxy模块的日志记录,记录请求详细信息

24w17b
---
- PRE-RELEASE: 此版本是v1.5.2的预发布版本,请勿在生产环境中使用
- FIX: 修正flag传入问题
- CHANGE: 去除/路径重定向,改为返回403,并记录对应请求日志
- CHANGE: 优化Proxy模块的日志记录,记录请求详细信息

24w17a
---
- PRE-RELEASE: 此版本是v1.5.2的预发布版本,请勿在生产环境中使用
- FIX: 初步修正flag传入问题,但仍有可能存在其他问题

v1.5.1
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: Bump github.com/imroc/req/v3 from 3.48.0 to 3.49.0 by @dependabot in https://github.com/WJQSERVER-STUDIO/ghproxy/pull/7
- ADD: 新增一键部署脚本,简化二进制文件部署流程

24w16a
---
- PRE-RELEASE: 此版本是v1.5.1的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: Bump github.com/imroc/req/v3 from 3.47.0 to 3.48.0 by @dependabot in https://github.com/WJQSERVER-STUDIO/ghproxy/pull/6
- ADD: 新增一键部署脚本,简化二进制文件部署流程

v1.5.0
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: 改进核心部分,即proxy模块的转发部分,对请求体处理与响应体处理进行优化
- CHANGE: 配置文件格式由yaml切换至toml,使其具备更好的可读性
- ADD: 黑白名单引入通配符支持,支持完全屏蔽或放行某个用户,例如`onwer/*`表示匹配`owner`的所有仓库
- ADD: 新增API模块,新增配置开关状态接口,以在前端指示功能状态
- CHANGE: 由于API变动,对前端进行相应调整
- ADD: 日志模块引入日志级别,排障更加直观
- CHANGE: 改进黑白名单机制,若禁用相关功能,则不对相关模块进行初始化

24w15d
---
- PRE-RELEASE: 此版本是v1.5.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- ADD: 新增API模块,新增配置开关状态接口,以在前端指示功能状态
- CHANGE: 由于API变动,对前端进行相应调整

24w15c
---
- PRE-RELEASE: 此版本是v1.5.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 改进核心部分,即proxy模块的转发部分,对请求体处理与响应体处理进行优化
- CHANGE: 改进黑白名单机制,若禁用相关功能,则不对对应模块进行初始化
- ADD: 黑白名单引入通配符支持,支持完全屏蔽或放行某个用户,例如`onwer/*`表示匹配`owner`的所有仓库
- ADD: 日志模块引入日志级别,排障更加直观

24w15b
---
- PRE-RELEASE: 此版本是v1.5.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- FIX: 修正24w15a版本的部分问题

24w15a
---
- PRE-RELEASE: 此版本是v1.5.0的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 将配置文件由yaml切换至toml

v1.4.3
---
- CHANGE: 优化代码结构,提升性能
- ADD: 新增命令行参数 `-cfg string` 用于指定配置文件路径
- CHANGE: 对二进制文件大小进行改进

24w14a
---
- PRE-RELEASE: 此版本是v1.4.3的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- ADD: 新增命令行参数 `-cfg string` 用于指定配置文件路径

v1.4.2
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: 初步引入ARM64架构支持
- CHANGE: 对Docker镜像构建进行优化，大幅减少镜像体积,从v1.4.0的`111 MB`,到v1.4.1的`58 MB`,再到v1.4.2的`28 MB`
- CHANGE: 切换至wjqserver/caddy:2.9.0-rc-alpine作为基础镜像

24w13c
---
- PRE-RELEASE: 此版本是v1.4.2的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 修正交叉编译问题

24w13b
---
- PRE-RELEASE: 此版本是v1.4.2的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 初步引入ARM64支持，但仍处于测试阶段
- CHANGE: 对Dockerfile进行优化，大幅减少镜像体积

24w13a
---
- PRE-RELEASE: 此版本是v1.4.2的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 更新相关依赖库

v1.4.1
---
- CHANGE: 优化代码结构,提升性能
- CHANGE: 引入Alpine Linux作为基础镜像,大幅减少Docker镜像体积
- FIX: 修正部分参数错误
- CHANGE: CGO_ENABLED=0

24w12c
---
- PRE-RELEASE: 此版本是v1.4.1的预发布版本,请勿在生产环境中使用
- CHANGE: 优化代码结构,提升性能
- CHANGE: 尝试在DEV版本引入Alpine Linux作为基础镜像,减少镜像体积

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
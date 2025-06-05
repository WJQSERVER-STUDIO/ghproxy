# Flag

> 弃用, 请转到 [GHProxy项目文档](https://wjqserver-docs.pages.dev/docs/ghproxy/)

GHProxy接受以下flag传入

```bash
root@root:/data/ghproxy$ ghproxy -h
  -c string
        config file path (default "/data/ghproxy/config/config.toml")
  -cfg value
        exit
  -h    show help message and exit
  -v    show version and exit
```

- `-c`
    类型: `string`
    默认值: `/data/ghproxy/config/config.toml`
    示例: `ghproxy -c /data/ghproxy/demo.toml`
- `-cfg`
    已弃用, 被`-c`替代
- `-h`
    显示帮助信息
- `-v`
    显示版本号

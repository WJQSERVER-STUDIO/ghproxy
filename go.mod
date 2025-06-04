module ghproxy

go 1.24.3

require (
	github.com/BurntSushi/toml v1.5.0
	github.com/WJQSERVER-STUDIO/httpc v0.5.1
	github.com/WJQSERVER-STUDIO/logger v1.7.3
	github.com/cloudwego/hertz v0.10.0
	github.com/hertz-contrib/http2 v0.1.8
	golang.org/x/net v0.40.0
	golang.org/x/time v0.11.0
)

require (
	github.com/WJQSERVER-STUDIO/go-utils/limitreader v0.0.2
	github.com/bytedance/sonic v1.13.3
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/wjqserver/modembed v0.0.1
)

require (
	github.com/WJQSERVER-STUDIO/go-utils/copyb v0.0.4 // indirect
	github.com/WJQSERVER-STUDIO/go-utils/log v0.0.3 // indirect
	github.com/bytedance/gopkg v0.1.2 // indirect
	github.com/bytedance/sonic/loader v0.2.4 // indirect
	github.com/cloudwego/base64x v0.1.5 // indirect
	github.com/cloudwego/gopkg v0.1.4 // indirect
	github.com/cloudwego/netpoll v0.7.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/nyaruka/phonenumbers v1.6.3 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	golang.org/x/arch v0.17.0 // indirect
	golang.org/x/exp v0.0.0-20250531010427-b6e5de432a8b // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)

replace github.com/nyaruka/phonenumbers => github.com/nyaruka/phonenumbers v1.6.1 // 1.6.3 has reflect leaking

//replace github.com/WJQSERVER-STUDIO/httpc v0.5.1 => /data/github/WJQSERVER-STUDIO/httpc
//replace github.com/WJQSERVER-STUDIO/logger v1.6.0 => /data/github/WJQSERVER-STUDIO/logger

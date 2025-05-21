package proxy

import (
	"fmt"
	"strings"
)

// BearerAuthParams 用于存放解析出的 Bearer 认证参数
type BearerAuthParams struct {
	Realm   string
	Service string
	Scope   string
}

// parseBearerWWWAuthenticateHeader 解析 Bearer 方案的 Www-Authenticate Header。
// 它期望格式为 'Bearer key1="value1",key2="value2",...'
// 并尝试将已知参数解析到 BearerAuthParams struct 中。
func parseBearerWWWAuthenticateHeader(headerValue string) (*BearerAuthParams, error) {
	if headerValue == "" {
		return nil, fmt.Errorf("header value is empty")
	}

	// 检查 Scheme 是否是 "Bearer"
	parts := strings.SplitN(headerValue, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, fmt.Errorf("invalid or non-bearer header format: got '%s'", headerValue)
	}
	paramsStr := parts[1]

	paramPairs := strings.Split(paramsStr, ",")
	tempMap := make(map[string]string)

	for _, pair := range paramPairs {
		trimmedPair := strings.TrimSpace(pair)
		keyValue := strings.SplitN(trimmedPair, "=", 2)
		if len(keyValue) != 2 {
			logWarning("Skipping malformed parameter '%s' in Www-Authenticate header: %s", pair, headerValue)
			continue
		}
		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = value[1 : len(value)-1]
		}
		tempMap[key] = value
	}

	//从 map 中提取值并填充到 struct
	authParams := &BearerAuthParams{}

	if realm, ok := tempMap["realm"]; ok {
		authParams.Realm = realm
	}
	if service, ok := tempMap["service"]; ok {
		authParams.Service = service
	}
	if scope, ok := tempMap["scope"]; ok {
		authParams.Scope = scope
	}

	return authParams, nil
}

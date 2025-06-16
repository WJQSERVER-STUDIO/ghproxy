package proxy

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"sync"

	"fmt"
	"html/template"
	"io/fs"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/infinite-iroha/touka"
)

func HandleError(c *touka.Context, message string) {
	ErrorPage(c, NewErrorWithStatusLookup(500, message))
	c.Errorf("%s %s %s %s %s Error: %v", c.ClientIP(), c.Request.Method, c.Request.URL.Path, c.UserAgent(), c.Request.Proto, message)
}

type GHProxyErrors struct {
	StatusCode   int
	StatusDesc   string
	StatusText   string
	HelpInfo     string
	ErrorMessage string
}

var (
	ErrInvalidURL = &GHProxyErrors{
		StatusCode: 400,
		StatusDesc: "Bad Request",
		StatusText: "无效请求",
		HelpInfo:   "请求的URL格式不正确，请检查后重试。",
	}
	ErrAuthHeaderUnavailable = &GHProxyErrors{
		StatusCode: 401,
		StatusDesc: "Unauthorized",
		StatusText: "认证失败",
		HelpInfo:   "缺少或无效的鉴权信息。",
	}
	ErrForbidden = &GHProxyErrors{
		StatusCode: 403,
		StatusDesc: "Forbidden",
		StatusText: "权限不足",
		HelpInfo:   "您没有权限访问此资源。",
	}
	ErrNotFound = &GHProxyErrors{
		StatusCode: 404,
		StatusDesc: "Not Found",
		StatusText: "页面未找到",
		HelpInfo:   "抱歉，您访问的页面不存在。",
	}
	ErrTooManyRequests = &GHProxyErrors{
		StatusCode: 429,
		StatusDesc: "Too Many Requests",
		StatusText: "请求过于频繁",
		HelpInfo:   "您的请求过于频繁，请稍后再试。",
	}
	ErrInternalServerError = &GHProxyErrors{
		StatusCode: 500,
		StatusDesc: "Internal Server Error",
		StatusText: "服务器内部错误",
		HelpInfo:   "服务器处理您的请求时发生错误，请稍后重试或联系管理员。",
	}
)

var statusErrorMap map[int]*GHProxyErrors

func init() {
	statusErrorMap = map[int]*GHProxyErrors{
		ErrInvalidURL.StatusCode:            ErrInvalidURL,
		ErrAuthHeaderUnavailable.StatusCode: ErrAuthHeaderUnavailable,
		ErrForbidden.StatusCode:             ErrForbidden,
		ErrNotFound.StatusCode:              ErrNotFound,
		ErrTooManyRequests.StatusCode:       ErrTooManyRequests,
		ErrInternalServerError.StatusCode:   ErrInternalServerError,
	}
}

func NewErrorWithStatusLookup(statusCode int, errMsg string) *GHProxyErrors {
	baseErr, found := statusErrorMap[statusCode]

	if found {
		return &GHProxyErrors{
			StatusCode:   baseErr.StatusCode,
			StatusDesc:   baseErr.StatusDesc,
			StatusText:   baseErr.StatusText,
			HelpInfo:     baseErr.HelpInfo,
			ErrorMessage: errMsg,
		}
	} else {
		return &GHProxyErrors{
			StatusCode:   statusCode,
			ErrorMessage: errMsg,
		}
	}
}

var errPagesFs fs.FS

func InitErrPagesFS(pages fs.FS) error {
	var err error
	errPagesFs, err = fs.Sub(pages, "pages/err")
	if err != nil {
		return err
	}
	return nil
}

type ErrorPageData struct {
	StatusCode   int
	StatusDesc   string
	StatusText   string
	HelpInfo     string
	ErrorMessage string
}

// ToCacheKey 为 ErrorPageData 生成一个唯一的 SHA256 字符串键。
// 使用 gob 序列化来确保结构体内容到字节序列的顺序一致性，然后计算哈希。
func (d ErrorPageData) ToCacheKey() (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(d)
	if err != nil {
		//logError("Failed to gob encode ErrorPageData for cache key: %v", err)
		return "", fmt.Errorf("failed to gob encode ErrorPageData for cache key: %w", err)
	}

	hasher := sha256.New()
	hasher.Write(buf.Bytes())
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func ErrPageUnwarper(errInfo *GHProxyErrors) ErrorPageData {
	return ErrorPageData{
		StatusCode:   errInfo.StatusCode,
		StatusDesc:   errInfo.StatusDesc,
		StatusText:   errInfo.StatusText,
		HelpInfo:     errInfo.HelpInfo,
		ErrorMessage: errInfo.ErrorMessage,
	}
}

// SizedLRUCache 实现了基于字节大小限制的 LRU 缓存。
// 它包装了 hashicorp/golang-lru/v2.Cache，并额外管理缓存的总字节大小。
type SizedLRUCache struct {
	cache        *lru.Cache[string, []byte]
	mu           sync.Mutex // 保护 currentBytes 字段
	maxBytes     int64      // 缓存的最大字节容量
	currentBytes int64      // 缓存当前占用的字节数
}

// NewSizedLRUCache 创建一个新的 SizedLRUCache 实例。
// 内部的 lru.Cache 的条目容量被设置为一个较大的值 (例如 10000)，
// 因为主要的逐出逻辑将由字节大小限制来控制。
func NewSizedLRUCache(maxBytes int64) (*SizedLRUCache, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("maxBytes must be positive")
	}

	c := &SizedLRUCache{
		maxBytes: maxBytes,
	}

	// 创建内部 LRU 缓存，并提供一个 OnEvictedFunc 回调函数。
	// 当内部 LRU 缓存因其自身的条目容量限制或 RemoveOldest 方法被调用而逐出条目时，
	// 此回调函数会被执行，从而更新 currentBytes。
	var err error
	c.cache, err = lru.NewWithEvict[string, []byte](10000, func(key string, value []byte) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.currentBytes -= int64(len(value))
		//logDebug("LRU evicted key: %s, size: %d, current total: %d", key, len(value), c.currentBytes)
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Get 从缓存中检索值。
func (c *SizedLRUCache) Get(key string) ([]byte, bool) {
	return c.cache.Get(key)
}

// Add 向缓存中添加或更新一个键值对，并在必要时执行逐出以满足字节限制。
func (c *SizedLRUCache) Add(key string, value []byte) {
	c.mu.Lock() // 保护 currentBytes 和逐出逻辑
	defer c.mu.Unlock()

	itemSize := int64(len(value))

	// 如果待添加的条目本身就大于缓存的最大容量，则不进行缓存。
	if itemSize > c.maxBytes {
		//c.Warnf("Item key %s (size %d) larger than cache max capacity %d. Not caching.", key, itemSize, c.maxBytes)
		return
	}

	// 如果键已存在，则首先从 currentBytes 中减去旧值的大小，并从内部 LRU 中移除旧条目。
	if oldVal, ok := c.cache.Get(key); ok {
		c.currentBytes -= int64(len(oldVal))
		c.cache.Remove(key)
		//logDebug("Key %s exists, removed old size %d. Current total: %d", key, len(oldVal), c.currentBytes)
	}

	// 主动逐出最旧的条目，直到有足够的空间容纳新条目。
	for c.currentBytes+itemSize > c.maxBytes && c.cache.Len() > 0 {
		_, _, existed := c.cache.RemoveOldest()
		if !existed {
			//c.Warnf("Attempted to remove oldest, but item not found.")
			break
		}
		//logDebug("Proactively evicted item (size %d) to free space. Current total: %d", len(oldVal), c.currentBytes)
	}

	// 添加新条目到内部 LRU 缓存。
	c.cache.Add(key, value)
	c.currentBytes += itemSize // 手动增加新条目的大小到 currentBytes。
	//logDebug("Item added: key %s, size: %d, current total: %d", key, itemSize, c.currentBytes)
}

const maxErrorPageCacheBytes = 512 * 1024 // 错误页面缓存的最大容量：512KB

var errorPageCache *SizedLRUCache

func init() {
	// 初始化 SizedLRUCache。
	var err error
	errorPageCache, err = NewSizedLRUCache(maxErrorPageCacheBytes)
	if err != nil {
		//	logError("Failed to initialize error page LRU cache: %v", err)
		panic(err)
	}
}

// parsedTemplateOnce 用于确保 HTML 模板只被解析一次。
var (
	parsedTemplateOnce sync.Once
	parsedTemplate     *template.Template
	parsedTemplateErr  error
)

// getParsedTemplate 用于获取缓存的解析后的 HTML 模板。
func getParsedTemplate() (*template.Template, error) {
	parsedTemplateOnce.Do(func() {
		tmplPath := "page.tmpl"
		// 确保 errPagesFs 已初始化。这要求在任何 ErrorPage 调用之前调用 InitErrPagesFS。
		if errPagesFs == nil {
			parsedTemplateErr = fmt.Errorf("errPagesFs not initialized. Call InitErrPagesFS first")
			return
		}
		parsedTemplate, parsedTemplateErr = template.ParseFS(errPagesFs, tmplPath)
		if parsedTemplateErr != nil {
			parsedTemplate = nil
		}
	})
	return parsedTemplate, parsedTemplateErr
}

// htmlTemplateRender 修改为使用缓存的模板。
func htmlTemplateRender(data interface{}) ([]byte, error) {
	tmpl, err := getParsedTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to get parsed template: %w", err)
	}
	if tmpl == nil {
		return nil, fmt.Errorf("template is nil after parsing")
	}

	// 创建一个 bytes.Buffer 用于存储渲染结果
	var buf bytes.Buffer

	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// 返回 buffer 的内容作为 []byte
	return buf.Bytes(), nil
}

func ErrorPage(c *touka.Context, errInfo *GHProxyErrors) {
	// 将 errInfo 转换为 ErrorPageData 结构体
	var err error
	var cacheKey string
	pageDataStruct := ErrPageUnwarper(errInfo)
	// 使用 ErrorPageData 生成一个唯一的 SHA256 缓存键
	cacheKey, err = pageDataStruct.ToCacheKey()
	if err != nil {
		c.Warnf("Failed to generate cache key for error page: %v", err)
		fallbackErrorJson(c, errInfo)
		return
	}

	// 检查生成的缓存键是否为空，这可能表示序列化或哈希计算失败

	if cacheKey == "" {
		c.JSON(errInfo.StatusCode, map[string]string{"error": errInfo.ErrorMessage})
		c.Warnf("Failed to generate cache key for error page: %v", errInfo)
		return
	}

	var pageData []byte

	// 尝试从缓存中获取页面数据
	if cachedPage, found := errorPageCache.Get(cacheKey); found {
		pageData = cachedPage
		c.Debugf("Serving error page from cache (Key: %s)", cacheKey)
	} else {
		// 如果不在缓存中，则渲染页面
		pageData, err = htmlTemplateRender(pageDataStruct)
		if err != nil {
			c.JSON(errInfo.StatusCode, map[string]string{"error": errInfo.ErrorMessage})
			c.Warnf("Failed to render error page for status %d (Key: %s): %v", errInfo.StatusCode, cacheKey, err)
			return
		}

		// 将渲染结果存入缓存
		errorPageCache.Add(cacheKey, pageData)
		c.Debugf("Cached error page (Key: %s, Size: %d bytes)", cacheKey, len(pageData))
	}

	c.Raw(errInfo.StatusCode, "text/html; charset=utf-8", pageData)
}

func fallbackErrorJson(c *touka.Context, errInfo *GHProxyErrors) {
	c.JSON(errInfo.StatusCode, map[string]string{"error": errInfo.ErrorMessage})
}

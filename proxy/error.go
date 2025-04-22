package proxy

import (
	"net/http"

	"github.com/WJQSERVER-STUDIO/go-utils/logger"
	"github.com/cloudwego/hertz/pkg/app"
)

// 日志模块
var (
	logw       = logger.Logw
	logDump    = logger.LogDump
	logDebug   = logger.LogDebug
	logInfo    = logger.LogInfo
	logWarning = logger.LogWarning
	logError   = logger.LogError
)

func HandleError(c *app.RequestContext, message string) {
	c.JSON(http.StatusInternalServerError, map[string]string{"error": message})
	logError(message)
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

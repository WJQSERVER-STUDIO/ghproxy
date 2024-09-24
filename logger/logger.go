// logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var logFile *os.File
var logger *log.Logger

// Init 初始化日志记录器，接受日志文件路径作为参数
func Init(logFilePath string) error {
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logger = log.New(logFile, "", 0) // 不使用默认前缀
	return nil
}

// Log 直接记录日志的函数，带有时间戳
func Log(customMessage string) {
	if logger != nil {
		timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700") // 使用自定义时间格式
		logger.Println(timestamp + " - " + customMessage)
	}
}

// Logw 用于格式化日志记录
func Logw(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...) // 格式化消息
	Log(message)                            // 记录日志
}

// Close 关闭日志文件
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

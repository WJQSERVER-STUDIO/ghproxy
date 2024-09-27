// logger/logger.go
package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logFile     *os.File
	logger      *log.Logger
	logChannel  = make(chan string, 100) // 创建一个缓冲通道
	quitChannel = make(chan struct{})    // 用于通知退出
)

// Init 初始化日志记录器，接受日志文件路径作为参数
func Init(logFilePath string) error {
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logger = log.New(logFile, "", 0) // 不使用默认前缀

	go logWorker() // 启动 goroutine 处理日志
	return nil
}

// logWorker 处理日志记录
func logWorker() {
	for {
		select {
		case msg := <-logChannel:
			timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
			logger.Println(timestamp + " - " + msg) // 写入日志
		case <-quitChannel:
			return // 退出 goroutine
		}
	}
}

// Log 直接记录日志的函数
func Log(customMessage string) {
	logChannel <- customMessage // 将日志消息发送到通道
}

// Logw 用于格式化日志记录
func Logw(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...) // 格式化消息
	Log(message)                            // 记录日志
}

// Close 关闭日志文件
func Close() {
	if logFile != nil {
		quitChannel <- struct{}{} // 通知日志 goroutine 退出
		if err := logFile.Close(); err != nil {
			Log("Error closing log file: " + err.Error()) // 记录关闭日志时的错误
		}
	}
}

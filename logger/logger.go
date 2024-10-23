package logger

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logw         = Logw
	logFile      *os.File
	logger       *log.Logger
	logChannel   = make(chan string, 100)
	quitChannel  = make(chan struct{})
	logFileMutex sync.Mutex
	logFilePath  = "/data/ghproxy/log/ghproxy.log"
)

// 初始化，接受日志文件路径作为参数
func Init(logFilePath_input string, maxLogsize int) error {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	var err error
	logFilePath = logFilePath_input
	logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	logger = log.New(logFile, "", 0)

	go logWorker()
	go monitorLogSize(logFilePath, maxLogsize)
	return nil
}

func logWorker() {
	for {
		select {
		case msg := <-logChannel:
			timestamp := time.Now().Format("02/Jan/2006:15:04:05 -0700")
			logger.Println(timestamp + " - " + msg)
		case <-quitChannel:
			return
		}
	}
}

func Log(customMessage string) {
	logChannel <- customMessage
}

// 格式化日志记录
func Logw(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Log(message)
}

// INFO
func LogInfo(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	output := fmt.Sprintf("[INFO] %s", message)
	Log(output)
}

// WARNING
func LogWarning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	output := fmt.Sprintf("[WARNING] %s", message)
	Log(output)
}

// ERROR
func LogError(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	Log(message)
}

// 关闭日志文件
func Close() {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if logFile != nil {
		quitChannel <- struct{}{}
		if err := logFile.Close(); err != nil {
			fmt.Printf("Error closing log file: %v", err)
		}
	}
}

func monitorLogSize(logFilePath string, maxLogsize int) {
	var maxLogsizeBytes int64 = int64(maxLogsize) * 1024 * 1024
	for {
		time.Sleep(120 * time.Minute) // 每120分钟检查一次日志文件大小
		logFileMutex.Lock()
		info, err := logFile.Stat()
		logFileMutex.Unlock()

		if err == nil && info.Size() > maxLogsizeBytes {
			if err := rotateLogFile(logFilePath); err != nil {
				logw("Log Rotation Failed: %s", err)
			}
		}
	}
}

func rotateLogFile(logFilePath string) error {
	logFileMutex.Lock()
	defer logFileMutex.Unlock()

	if logFile != nil {
		if err := logFile.Close(); err != nil {
			logw("Error closing log file for rotation: %v", err)
		}
	}

	// 打开当前日志文件
	logFile, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("failed to open log file: %s, error: %w", logFilePath, err)
	}
	defer logFile.Close()

	newLogFilePath := logFilePath + "-" + time.Now().Format("20060102-150405") + ".tar.gz"
	outFile, err := os.Create(newLogFilePath)
	if err != nil {
		return fmt.Errorf("failed to create gz file: %s, error: %w", newLogFilePath, err)
	}
	defer outFile.Close()

	gzWriter, err := gzip.NewWriterLevel(outFile, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("failed to create gz writer: %w", err)
	}
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	logFileStat, err := logFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat log file: %s, error: %w", logFilePath, err)
	}

	logFileHeader := &tar.Header{
		Name:    filepath.Base(logFilePath),
		Size:    logFileStat.Size(),
		Mode:    0644,
		ModTime: logFileStat.ModTime(),
	}

	if err := tarWriter.WriteHeader(logFileHeader); err != nil {
		return fmt.Errorf("failed to write log file header: %s, error: %w", logFilePath, err)
	}

	if _, err := io.Copy(tarWriter, logFile); err != nil {
		return fmt.Errorf("failed to copy log file: %s, error: %w", logFilePath, err)
	}

	if err := os.Truncate(logFilePath, 0); err != nil {
		return fmt.Errorf("failed to truncate log file: %s, error: %w", logFilePath, err)
	}

	// 重新打开日志文件
	logFile, err = os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to reopen log file: %s, error: %w", logFilePath, err)
	}
	logger.SetOutput(logFile)

	return nil
}

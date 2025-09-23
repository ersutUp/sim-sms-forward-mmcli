// Package logger 提供日志框架，支持文件输出、滚动存储和自动清理
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// 设置中国标准时间(CST, UTC+8)
var cstZone = time.FixedZone("CST", 8*3600)

// Logger 日志记录器
type Logger struct {
	infoLogger  *log.Logger // 信息日志记录器
	errorLogger *log.Logger // 错误日志记录器
	logDir      string      // 日志目录
	logFile     *os.File    // 当前日志文件
	lastDate    string      // 上次记录日志的日期
}

// LogLevel 日志级别
type LogLevel int

const (
	INFO LogLevel = iota
	ERROR
)

// 全局日志实例
var globalLogger *Logger

// Init 初始化日志系统
// 参数: logDir - 日志文件存储目录
// 返回: 初始化错误
func Init(logDir string) error {
	var err error
	globalLogger, err = NewLogger(logDir)
	if err != nil {
		return fmt.Errorf("初始化日志系统失败: %v", err)
	}

	// 启动日志清理协程
	go globalLogger.startLogCleaner()

	return nil
}

// NewLogger 创建新的日志记录器
// 参数: logDir - 日志文件存储目录
// 返回: Logger实例和可能的错误
func NewLogger(logDir string) (*Logger, error) {
	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}

	logger := &Logger{
		logDir: logDir,
	}

	// 初始化日志文件
	if err := logger.rotateLogFile(); err != nil {
		return nil, fmt.Errorf("初始化日志文件失败: %v", err)
	}

	return logger, nil
}

// rotateLogFile 轮转日志文件（按日期）
func (l *Logger) rotateLogFile() error {
	currentDate := time.Now().In(cstZone).Format("2006-01-02")

	// 如果是同一天，不需要轮转
	if l.lastDate == currentDate && l.logFile != nil {
		return nil
	}

	// 关闭当前日志文件
	if l.logFile != nil {
		l.logFile.Close()
	}

	// 创建新的日志文件
	logFilePath := filepath.Join(l.logDir, fmt.Sprintf("sms-forward-%s.log", currentDate))
	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("创建日志文件失败: %v", err)
	}

	l.logFile = file
	l.lastDate = currentDate

	// 创建多输出写入器（同时输出到控制台和文件）
	multiWriter := io.MultiWriter(os.Stdout, file)
	errorWriter := io.MultiWriter(os.Stderr, file)

	// 初始化日志记录器，使用自定义时间格式（24小时制，+8时区）
	l.infoLogger = log.New(multiWriter, "[INFO] ", 0)
	l.errorLogger = log.New(errorWriter, "[ERROR] ", 0)

	return nil
}

// formatTime 格式化时间为中国标准时间24小时制
func formatTime() string {
	return time.Now().In(cstZone).Format("2006/01/02 15:04:05")
}

// getCaller 获取调用者信息
// skip: 跳过的调用栈层数
// 返回: 文件名:行号 格式的字符串
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	// 只保留文件名，不要完整路径
	return fmt.Sprintf("%s:%d", filepath.Base(file), line)
}

// Info 记录信息日志
func (l *Logger) Info(v ...interface{}) {
	l.checkRotation()
	caller := getCaller(2) // 跳过当前函数和调用此函数的包装函数
	message := fmt.Sprint(v...)
	l.infoLogger.Printf("%s %s: %s", formatTime(), caller, message)
}

// Infof 记录格式化信息日志
func (l *Logger) Infof(format string, v ...interface{}) {
	l.checkRotation()
	caller := getCaller(2) // 跳过当前函数和调用此函数的包装函数
	message := fmt.Sprintf(format, v...)
	l.infoLogger.Printf("%s %s: %s", formatTime(), caller, message)
}

// Error 记录错误日志
func (l *Logger) Error(v ...interface{}) {
	l.checkRotation()
	caller := getCaller(2) // 跳过当前函数和调用此函数的包装函数
	message := fmt.Sprint(v...)
	l.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
}

// Errorf 记录格式化错误日志
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.checkRotation()
	caller := getCaller(2) // 跳过当前函数和调用此函数的包装函数
	message := fmt.Sprintf(format, v...)
	l.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
}

// Fatal 记录致命错误日志并退出程序
func (l *Logger) Fatal(v ...interface{}) {
	l.checkRotation()
	caller := getCaller(2) // 跳过当前函数和调用此函数的包装函数
	message := fmt.Sprint(v...)
	l.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
	os.Exit(1)
}

// Fatalf 记录格式化致命错误日志并退出程序
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.checkRotation()
	caller := getCaller(2) // 跳过当前函数和调用此函数的包装函数
	message := fmt.Sprintf(format, v...)
	l.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
	os.Exit(1)
}

// checkRotation 检查是否需要轮转日志文件
func (l *Logger) checkRotation() {
	currentDate := time.Now().In(cstZone).Format("2006-01-02")
	if l.lastDate != currentDate {
		if err := l.rotateLogFile(); err != nil {
			// 如果轮转失败，记录到标准错误输出
			fmt.Fprintf(os.Stderr, "日志文件轮转失败: %v\n", err)
		}
	}
}

// startLogCleaner 启动日志清理协程
func (l *Logger) startLogCleaner() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := l.cleanOldLogs(); err != nil {
				l.Errorf("清理旧日志文件失败: %v", err)
			}
		}
	}
}

// cleanOldLogs 清理3天前的日志文件
func (l *Logger) cleanOldLogs() error {
	// 计算3天前的日期（使用中国标准时间）
	cutoffDate := time.Now().In(cstZone).AddDate(0, 0, -3)
	cutoffDateStr := cutoffDate.Format("2006-01-02")

	// 读取日志目录
	files, err := os.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("读取日志目录失败: %v", err)
	}

	// 查找需要删除的日志文件
	var filesToDelete []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// 检查是否是日志文件格式: sms-forward-YYYY-MM-DD.log
		if strings.HasPrefix(file.Name(), "sms-forward-") && strings.HasSuffix(file.Name(), ".log") {
			// 提取日期部分
			fileName := file.Name()
			if len(fileName) >= 25 { // "sms-forward-YYYY-MM-DD.log" 长度为25
				dateStr := fileName[12:22] // 提取 YYYY-MM-DD 部分
				if dateStr < cutoffDateStr {
					filesToDelete = append(filesToDelete, filepath.Join(l.logDir, fileName))
				}
			}
		}
	}

	// 删除旧文件
	for _, filePath := range filesToDelete {
		if err := os.Remove(filePath); err != nil {
			l.Errorf("删除旧日志文件失败 %s: %v", filePath, err)
		} else {
			l.Infof("已删除旧日志文件: %s", filePath)
		}
	}

	if len(filesToDelete) > 0 {
		l.Infof("清理完成，删除了 %d 个旧日志文件", len(filesToDelete))
	}

	return nil
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// 全局日志函数，方便在其他包中使用

// Info 记录信息日志
func Info(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.checkRotation()
		caller := getCaller(2) // 跳过当前函数和调用者
		message := fmt.Sprint(v...)
		globalLogger.infoLogger.Printf("%s %s: %s", formatTime(), caller, message)
	} else {
		log.Println("[INFO]", fmt.Sprint(v...))
	}
}

// Infof 记录格式化信息日志
func Infof(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.checkRotation()
		caller := getCaller(2) // 跳过当前函数和调用者
		message := fmt.Sprintf(format, v...)
		globalLogger.infoLogger.Printf("%s %s: %s", formatTime(), caller, message)
	} else {
		log.Printf("[INFO] "+format, v...)
	}
}

// Error 记录错误日志
func Error(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.checkRotation()
		caller := getCaller(2) // 跳过当前函数和调用者
		message := fmt.Sprint(v...)
		globalLogger.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
	} else {
		log.Println("[ERROR]", fmt.Sprint(v...))
	}
}

// Errorf 记录格式化错误日志
func Errorf(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.checkRotation()
		caller := getCaller(2) // 跳过当前函数和调用者
		message := fmt.Sprintf(format, v...)
		globalLogger.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
	} else {
		log.Printf("[ERROR] "+format, v...)
	}
}

// Fatal 记录致命错误日志并退出程序
func Fatal(v ...interface{}) {
	if globalLogger != nil {
		globalLogger.checkRotation()
		caller := getCaller(2) // 跳过当前函数和调用者
		message := fmt.Sprint(v...)
		globalLogger.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
		os.Exit(1)
	} else {
		log.Fatal("[FATAL]", fmt.Sprint(v...))
	}
}

// Fatalf 记录格式化致命错误日志并退出程序
func Fatalf(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.checkRotation()
		caller := getCaller(2) // 跳过当前函数和调用者
		message := fmt.Sprintf(format, v...)
		globalLogger.errorLogger.Printf("%s %s: %s", formatTime(), caller, message)
		os.Exit(1)
	} else {
		log.Fatalf("[FATAL] "+format, v...)
	}
}

// GetLogFiles 获取所有日志文件列表（按日期排序）
func GetLogFiles() ([]string, error) {
	if globalLogger == nil {
		return nil, fmt.Errorf("日志系统未初始化")
	}

	files, err := os.ReadDir(globalLogger.logDir)
	if err != nil {
		return nil, fmt.Errorf("读取日志目录失败: %v", err)
	}

	var logFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasPrefix(file.Name(), "sms-forward-") && strings.HasSuffix(file.Name(), ".log") {
			logFiles = append(logFiles, filepath.Join(globalLogger.logDir, file.Name()))
		}
	}

	// 按文件名排序（实际上就是按日期排序）
	sort.Strings(logFiles)
	return logFiles, nil
}

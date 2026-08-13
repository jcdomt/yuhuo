package yuhuo

import (
	"fmt"
	"time"
)

// 系统日志记录器

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

type DefaultLogger struct{}

func (l *DefaultLogger) defaultLogHandler(level string, args ...interface{}) string {
	// 日志通用头部消息
	// 目前格式为 "[LEVEL] [时间戳] [消息]"
	return fmt.Sprintf("[%s] [%s] %s", level, time.Now().Format("2006-01-02 15:04:05"), fmt.Sprint(args...))
}

func (l *DefaultLogger) Debug(args ...interface{}) {
	fmt.Println(l.defaultLogHandler("DEBUG", args...))
}

func (l *DefaultLogger) Info(args ...interface{}) {
	fmt.Println(l.defaultLogHandler("INFO", args...))
}

func (l *DefaultLogger) Warn(args ...interface{}) {
	fmt.Println(l.defaultLogHandler("WARN", args...))
}

func (l *DefaultLogger) Error(args ...interface{}) {
	fmt.Println(l.defaultLogHandler("ERROR", args...))
}

func (l *DefaultLogger) Fatal(args ...interface{}) {
	fmt.Println(l.defaultLogHandler("FATAL", args...))
	// 终止程序执行
	panic(fmt.Sprint(args...))
}

func GetDefaultLogger() *DefaultLogger {
	return &DefaultLogger{}
}

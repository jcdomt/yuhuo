package yuhuo

import (
	"fmt"
	"time"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// Logger 定义应用使用的日志能力。
type Logger = requestcontext.Logger

// DefaultLogger 是默认的标准输出日志器。
type DefaultLogger struct{}

func (l *DefaultLogger) defaultLogHandler(level string, args ...interface{}) string {
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
	panic(fmt.Sprint(args...))
}
func GetDefaultLogger() *DefaultLogger { return &DefaultLogger{} }

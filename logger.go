package yuhuo

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// Logger 定义应用使用的日志能力。
type Logger = requestcontext.Logger

const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
	LogLevelFatal = "FATAL"
)

const (
	levelDebug = iota + 1
	levelInfo
	levelWarn
	levelError
	levelFatal
)

// DefaultLogger 是默认的标准输出日志器，默认级别为 DEBUG。
type DefaultLogger struct {
	level atomic.Int32
}

func (l *DefaultLogger) defaultLogHandler(level string, args ...interface{}) string {
	return fmt.Sprintf("[%s]\t[%s]\t%s", level, time.Now().Format("2006-01-02 15:04:05"), fmt.Sprint(args...))
}
func (l *DefaultLogger) Debug(args ...interface{}) {
	if l.level.Load() > levelDebug {
		return
	}
	fmt.Println(l.defaultLogHandler("DEBUG", args...))
}
func (l *DefaultLogger) Info(args ...interface{}) {
	if l.level.Load() > levelInfo {
		return
	}
	fmt.Println(l.defaultLogHandler("INFO", args...))
}
func (l *DefaultLogger) Warn(args ...interface{}) {
	if l.level.Load() > levelWarn {
		return
	}
	fmt.Println(l.defaultLogHandler("WARN", args...))
}
func (l *DefaultLogger) Error(args ...interface{}) {
	if l.level.Load() > levelError {
		return
	}
	fmt.Println(l.defaultLogHandler("ERROR", args...))
}
func (l *DefaultLogger) Fatal(args ...interface{}) {
	if l.level.Load() > levelFatal {
		return
	}
	fmt.Println(l.defaultLogHandler("FATAL", args...))
	panic(fmt.Sprint(args...))
}

// SetLevel 设置日志级别。级别名不区分大小写，未知级别将被忽略。
func (l *DefaultLogger) SetLevel(level string) {
	switch strings.ToUpper(level) {
	case LogLevelDebug:
		l.level.Store(levelDebug)
	case LogLevelInfo:
		l.level.Store(levelInfo)
	case LogLevelWarn:
		l.level.Store(levelWarn)
	case LogLevelError:
		l.level.Store(levelError)
	case LogLevelFatal:
		l.level.Store(levelFatal)
	}
}
func (l *DefaultLogger) Log(args ...interface{}) {
	l.Info(args...)
}

// GetDefaultLogger 返回默认的标准输出日志器。
// 注意：每次调用都会返回新的实例，若要调整已使用日志器的级别，请通过 Application.Logger() 获取。
func GetDefaultLogger() *DefaultLogger {
	logger := &DefaultLogger{}
	logger.level.Store(levelDebug)
	return logger
}

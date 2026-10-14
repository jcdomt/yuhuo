package context

// Logger 是请求上下文能够使用的最小日志能力
type Logger interface {
	// Debug 输出调试级别日志
	Debug(args ...interface{})
	// Info 输出信息级别日志
	Info(args ...interface{})
	// Warn 输出警告级别日志
	Warn(args ...interface{})
	// Error 输出错误级别日志
	Error(args ...interface{})
	// Fatal 输出致命级别日志并触发 panic
	Fatal(args ...interface{})

	// SetLevel 设置日志级别
	SetLevel(level string)
}

// Application 描述上下文所需的应用能力，避免 context 反向依赖根包
type Application interface {
	// Logger 返回应用的日志记录器
	Logger() Logger
}

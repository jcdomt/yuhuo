package context

// Logger 是请求上下文能够使用的最小日志能力。
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// Application 描述上下文所需的应用能力，避免 context 反向依赖根包。
type Application interface {
	Logger() Logger
}

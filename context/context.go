package context

import (
	"context"
	"net/http"
	"net/url"

	"github.com/jcdomt/yuhuo/router"
)

// Context 封装一次 HTTP 请求的状态、上下文值和处理链
type Context struct {
	// Ctx 是标准库上下文，用于传递请求级的键值
	Ctx context.Context

	request  *http.Request
	response http.ResponseWriter

	statusCode int
	committed  bool
	handlers   []HandlerFunc
	index      int
	aborted    bool

	query      url.Values
	body       []byte
	formParsed bool

	application Application
}

// NewContext	为请求创建上下文
//
// param:
//   - r	HTTP 请求
//   - w	响应写入器
//   - application	所属应用
// return:
//   - 请求上下文
func NewContext(r *http.Request, w http.ResponseWriter, application Application) *Context {
	return &Context{Ctx: r.Context(), request: r, response: w, statusCode: http.StatusOK, index: -1, application: application}
}

// SetHandlers	设置当前请求的处理链
//
// param:
//   - handlers	处理函数链
func (ctx *Context) SetHandlers(handlers []HandlerFunc) {
	ctx.handlers = append([]HandlerFunc(nil), handlers...)
}

// Param	获取路由参数
//
// param:
//   - name	路由参数名
// return:
//   - 路由参数值
func (ctx *Context) Param(name string) string { return router.Param(ctx.request, name) }

// Request	返回原始 HTTP 请求
//
// return:
//   - 原始 HTTP 请求
func (ctx *Context) Request() *http.Request { return ctx.request }

// Response	返回原始响应写入器
//
// return:
//   - 原始响应写入器
func (ctx *Context) Response() http.ResponseWriter { return ctx.response }

// Set	在请求上下文中保存值
//
// param:
//   - key	键
//   - value	值
func (ctx *Context) Set(key string, value interface{}) {
	ctx.Ctx = context.WithValue(ctx.Ctx, key, value)
}

// Get	获取请求上下文中的值
//
// param:
//   - key	键
// return:
//   - 对应的值，不存在时返回 nil
func (ctx *Context) Get(key string) interface{} { return ctx.Ctx.Value(key) }

// Next	执行当前处理链中剩余的处理函数
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < len(ctx.handlers) && !ctx.aborted {
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

// Abort	停止执行后续处理函数
func (ctx *Context) Abort() { ctx.aborted = true }

// IsAborted	判断处理链是否已停止
//
// return:
//   - 处理链是否已停止
func (ctx *Context) IsAborted() bool { return ctx.aborted }

// AbortWithStatus	写入状态码并停止处理链
//
// param:
//   - statusCode	状态码
func (ctx *Context) AbortWithStatus(statusCode int) { ctx.SetStatus(statusCode); ctx.Abort() }

// InternalServerError	快速触发一次 Internal Server Error 响应并停止处理链
func (ctx *Context) InternalServerError() {
	ctx.AbortWithStatus(http.StatusInternalServerError)
	ctx.JSON(map[string]interface{}{"error": "Internal Server Error", "code": 500})
}

// SetStatus	设置响应状态码并提交响应头
//
// param:
//   - statusCode	状态码
func (ctx *Context) SetStatus(statusCode int) {
	ctx.writeHeader(statusCode)
}

// writeHeader	提交响应头，重复调用会被忽略
//
// param:
//   - statusCode	状态码
func (ctx *Context) writeHeader(statusCode int) {
	if ctx.committed {
		return
	}
	ctx.committed = true
	ctx.statusCode = statusCode
	ctx.response.WriteHeader(statusCode)
}

// IsCommitted	判断响应头是否已写出
//
// return:
//   - 响应头是否已写出
func (ctx *Context) IsCommitted() bool { return ctx.committed }

// Status	返回当前响应状态码
//
// return:
//   - 当前响应状态码
func (ctx *Context) Status() int { return ctx.statusCode }

// GetLogger	返回当前应用的日志记录器
//
// return:
//   - 日志记录器
func (ctx *Context) GetLogger() Logger { return ctx.application.Logger() }

// Logger	是 GetLogger 的别名，返回当前应用的日志记录器
//
// return:
//   - 日志记录器
func (ctx *Context) Logger() Logger { return ctx.application.Logger() }

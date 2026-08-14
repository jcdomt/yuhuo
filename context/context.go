package context

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/jcdomt/yuhuo/router"
)

// Context 封装一次 HTTP 请求的状态、上下文值和处理链。
type Context struct {
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

// NewContext 为请求创建上下文。
func NewContext(r *http.Request, w http.ResponseWriter, application Application) *Context {
	return &Context{Ctx: r.Context(), request: r, response: w, statusCode: http.StatusOK, index: -1, application: application}
}

// SetHandlers 设置当前请求的处理链。
func (ctx *Context) SetHandlers(handlers []HandlerFunc) {
	ctx.handlers = append([]HandlerFunc(nil), handlers...)
}

// Param 获取路由参数。
func (ctx *Context) Param(name string) string { return router.Param(ctx.request, name) }

// Request 返回原始 HTTP 请求。
func (ctx *Context) Request() *http.Request { return ctx.request }

// Response 返回原始响应写入器。
func (ctx *Context) Response() http.ResponseWriter { return ctx.response }

// Set 在请求上下文中保存值。
func (ctx *Context) Set(key string, value interface{}) {
	ctx.Ctx = context.WithValue(ctx.Ctx, key, value)
}

// Get 获取请求上下文中的值。
func (ctx *Context) Get(key string) interface{} { return ctx.Ctx.Value(key) }

// Next 执行当前处理链中剩余的处理函数。
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < len(ctx.handlers) && !ctx.aborted {
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

// Abort 停止执行后续处理函数。
func (ctx *Context) Abort() { ctx.aborted = true }

// IsAborted 判断处理链是否已停止。
func (ctx *Context) IsAborted() bool { return ctx.aborted }

// AbortWithStatus 写入状态码并停止处理链。
func (ctx *Context) AbortWithStatus(statusCode int) { ctx.SetStatus(statusCode); ctx.Abort() }

// 快速出发一次 Internal Server Error 响应并停止处理链。
func (ctx *Context) InternalServerError() {
	ctx.AbortWithStatus(http.StatusInternalServerError)
	ctx.JSON(map[string]interface{}{"error": "Internal Server Error", "code": 500})
}

// SetStatus 设置响应状态码并提交响应头。
func (ctx *Context) SetStatus(statusCode int) {
	ctx.writeHeader(statusCode)
}

// writeHeader 提交响应头，重复调用会被忽略。
func (ctx *Context) writeHeader(statusCode int) {
	if ctx.committed {
		return
	}
	ctx.committed = true
	ctx.statusCode = statusCode
	ctx.response.WriteHeader(statusCode)
}

// IsCommitted 判断响应头是否已写出。
func (ctx *Context) IsCommitted() bool { return ctx.committed }

// Status 返回当前响应状态码。
func (ctx *Context) Status() int { return ctx.statusCode }

// JSON 将数据编码为 JSON 响应。
func (ctx *Context) JSON(data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	ctx.response.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.writeHeader(ctx.statusCode)
	_, _ = ctx.response.Write(dataBytes)
}

// 带状态码的返回
func (ctx *Context) JSONs(statusCode int, data interface{}) {
	ctx.SetStatus(statusCode)
	ctx.JSON(data)
}

// GetLogger 返回当前应用的日志记录器。
func (ctx *Context) GetLogger() Logger { return ctx.application.Logger() }

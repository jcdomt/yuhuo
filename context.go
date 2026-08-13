package yuhuo

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jcdomt/yuhuo/router"
)

type Context struct {
	Ctx context.Context

	request  *http.Request
	response http.ResponseWriter

	statusCode int
	handlers   []HandlerFunc
	index      int
	aborted    bool

	application *Application
}

// 在请求一开始，生成 Context 对象，并将其存储在请求的上下文中，以便在后续处理中使用
func NewContext(r *http.Request, w http.ResponseWriter, application *Application) *Context {
	ctx := &Context{
		Ctx:      r.Context(),
		request:  r,
		response: w,

		statusCode: http.StatusOK,
		index:      -1,

		application: application,
	}
	return ctx
}

// Param 用于获取路由参数的值。它从请求的上下文中提取参数映射，并返回指定参数名称的值。
func (ctx *Context) Param(name string) string {
	return router.Param(ctx.request, name)
}

// Request 返回原始的 HTTP 请求对象
func (ctx *Context) Request() *http.Request {
	return ctx.request
}

// Response 返回原始的 HTTP 响应写入器
func (ctx *Context) Response() http.ResponseWriter {
	return ctx.response
}

func (ctx *Context) Set(key string, value interface{}) {
	ctx.Ctx = context.WithValue(ctx.Ctx, key, value)
}

func (ctx *Context) Get(key string) interface{} {
	return ctx.Ctx.Value(key)
}

// Next executes the remaining middlewares and the final route handler.
func (ctx *Context) Next() {
	ctx.index++
	for ctx.index < len(ctx.handlers) && !ctx.aborted {
		ctx.handlers[ctx.index](ctx)
		ctx.index++
	}
}

// Abort prevents the remaining handlers in the current chain from running.
func (ctx *Context) Abort() {
	ctx.aborted = true
}

// IsAborted reports whether this request's handler chain has been stopped.
func (ctx *Context) IsAborted() bool {
	return ctx.aborted
}

// AbortWithStatus stops the handler chain after writing statusCode.
func (ctx *Context) AbortWithStatus(statusCode int) {
	ctx.SetStatus(statusCode)
	ctx.Abort()
}

// SetStatus 设置 HTTP 响应状态码
func (ctx *Context) SetStatus(statusCode int) {
	ctx.statusCode = statusCode
	ctx.response.WriteHeader(statusCode)
}

// Status 返回当前的 HTTP 响应状态码
func (ctx *Context) Status() int {
	return ctx.statusCode
}

// JSON 用于返回一个 JSON 响应
func (ctx *Context) JSON(data interface{}) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.response.Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.response.WriteHeader(ctx.statusCode)
	_, _ = ctx.response.Write(dataBytes)
}

// 获取到系统的日志记录器
func (ctx *Context) GetLogger() Logger {
	return ctx.application.logger
}

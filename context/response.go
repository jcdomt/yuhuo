package context

import (
	"encoding/json"
	"net/http"
)

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

// String 以纯文本格式返回字符串。
func (ctx *Context) String(data string) {
	ctx.writeString("text/plain; charset=utf-8", data)
}

func (ctx *Context) Strings(statusCode int, data string) {
	ctx.SetStatus(statusCode)
	ctx.String(data)
}

// HTML 以 HTML 格式返回字符串。
func (ctx *Context) HTML(html string) {
	ctx.writeString("text/html; charset=utf-8", html)
}

func (ctx *Context) HTMLs(statusCode int, html string) {
	ctx.SetStatus(statusCode)
	ctx.HTML(html)
}

func (ctx *Context) writeString(contentType, data string) {
	ctx.response.Header().Set("Content-Type", contentType)
	ctx.writeHeader(ctx.statusCode)
	_, _ = ctx.response.Write([]byte(data))
}

// Data 以字节流返回原始数据。
func (ctx *Context) Data(data []byte) {
	ctx.response.Header().Set("Content-Type", "application/octet-stream")
	ctx.writeHeader(ctx.statusCode)
	_, _ = ctx.response.Write(data)
}

func (ctx *Context) Datas(statusCode int, data []byte) {
	ctx.SetStatus(statusCode)
	ctx.Data(data)
}

// Redirect 以 302 状态码重定向到指定地址。
func (ctx *Context) Redirect(location string) {
	ctx.RedirectWithStatus(http.StatusFound, location)
}

// RedirectWithStatus 以指定状态码重定向到指定地址。
func (ctx *Context) RedirectWithStatus(statusCode int, location string) {
	ctx.response.Header().Set("Location", location)
	ctx.writeHeader(statusCode)
}

// File 将磁盘文件作为响应返回。
func (ctx *Context) File(path string) {
	http.ServeFile(ctx.response, ctx.request, path)
}

func (ctx *Context) Header(key, value string) {
	ctx.response.Header().Set(key, value)
}

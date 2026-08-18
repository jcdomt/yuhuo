package context

import (
	"encoding/json"
	"net/http"
)

// JSON	将数据编码为 JSON 响应
//
// param:
//   - data	要编码为 JSON 的数据
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

// JSONs	以指定状态码返回 JSON 响应
//
// param:
//   - statusCode	响应状态码
//   - data	要编码为 JSON 的数据
func (ctx *Context) JSONs(statusCode int, data interface{}) {
	ctx.SetStatus(statusCode)
	ctx.JSON(data)
}

// String	以纯文本格式返回字符串
//
// param:
//   - data	文本内容
func (ctx *Context) String(data string) {
	ctx.writeString("text/plain; charset=utf-8", data)
}

// Strings	以指定状态码返回纯文本响应
//
// param:
//   - statusCode	响应状态码
//   - data	文本内容
func (ctx *Context) Strings(statusCode int, data string) {
	ctx.SetStatus(statusCode)
	ctx.String(data)
}

// HTML	以 HTML 格式返回字符串
//
// param:
//   - html	HTML 内容
func (ctx *Context) HTML(html string) {
	ctx.writeString("text/html; charset=utf-8", html)
}

// HTMLs	以指定状态码返回 HTML 响应
//
// param:
//   - statusCode	响应状态码
//   - html	HTML 内容
func (ctx *Context) HTMLs(statusCode int, html string) {
	ctx.SetStatus(statusCode)
	ctx.HTML(html)
}

// writeString	写入指定 Content-Type 的文本响应
//
// param:
//   - contentType	响应 Content-Type
//   - data	文本内容
func (ctx *Context) writeString(contentType, data string) {
	ctx.response.Header().Set("Content-Type", contentType)
	ctx.writeHeader(ctx.statusCode)
	_, _ = ctx.response.Write([]byte(data))
}

// Data	以字节流返回原始数据
//
// param:
//   - data	原始数据
func (ctx *Context) Data(data []byte) {
	ctx.response.Header().Set("Content-Type", "application/octet-stream")
	ctx.writeHeader(ctx.statusCode)
	_, _ = ctx.response.Write(data)
}

// Datas	以指定状态码返回原始数据
//
// param:
//   - statusCode	响应状态码
//   - data	原始数据
func (ctx *Context) Datas(statusCode int, data []byte) {
	ctx.SetStatus(statusCode)
	ctx.Data(data)
}

// Redirect	以 302 状态码重定向到指定地址
//
// param:
//   - location	重定向地址
func (ctx *Context) Redirect(location string) {
	ctx.RedirectWithStatus(http.StatusFound, location)
}

// RedirectWithStatus	以指定状态码重定向到指定地址
//
// param:
//   - statusCode	响应状态码
//   - location	重定向地址
func (ctx *Context) RedirectWithStatus(statusCode int, location string) {
	ctx.response.Header().Set("Location", location)
	ctx.writeHeader(statusCode)
}

// File	将磁盘文件作为响应返回
//
// param:
//   - path	文件路径
func (ctx *Context) File(path string) {
	http.ServeFile(ctx.response, ctx.request, path)
}

// Header	设置单个响应头
//
// param:
//   - key	响应头名称
//   - value	响应头值
func (ctx *Context) Header(key, value string) {
	ctx.response.Header().Set(key, value)
}

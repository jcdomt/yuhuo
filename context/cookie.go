package context

import "net/http"

// Cookie 是 http.Cookie 的兼容别名
type Cookie = http.Cookie

// SetCookie	设置响应 Cookie
//
// param:
//   - cookie	Cookie 信息，为 nil 时忽略
func (ctx *Context) SetCookie(cookie *http.Cookie) {
	if cookie == nil {
		return
	}
	http.SetCookie(ctx.response, cookie)
}

// GetCookie	获取请求中指定名称的 Cookie
//
// param:
//   - name	Cookie 名称
// return:
//   - Cookie 信息，不存在时返回错误
func (ctx *Context) GetCookie(name string) (*http.Cookie, error) {
	return ctx.request.Cookie(name)
}

// Cookie	获取请求中指定名称的 Cookie 值，不存在时返回空字符串
//
// param:
//   - name	Cookie 名称
// return:
//   - Cookie 值
func (ctx *Context) Cookie(name string) string {
	cookie, err := ctx.request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// ClearCookie	通过设置过期时间删除指定名称的 Cookie
//
// param:
//   - name	Cookie 名称
func (ctx *Context) ClearCookie(name string) {
	http.SetCookie(ctx.response, &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1})
}

package yuhuo

import (
	"net/http"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// Context 是 context 包类型的根包兼容别名
type Context = requestcontext.Context

// ControllerFunc 是 context.ControllerFunc 的根包兼容别名
type ControllerFunc = requestcontext.ControllerFunc

// NewContext	创建绑定到当前应用的请求上下文
// 用于统一创建请求上下文，方便在应用中使用
//
// param:
//   - r	HTTP 请求
//   - w	响应写入器
//   - application	所属应用实例
// return:
//   - 请求上下文
func NewContext(r *http.Request, w http.ResponseWriter, application *Application) *Context {
	return requestcontext.NewContext(r, w, application)
}

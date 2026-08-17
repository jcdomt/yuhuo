package yuhuo

import (
	"net/http"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// Context 是 context 包类型的根包兼容别名。
type Context = requestcontext.Context

type ControllerFunc = requestcontext.ControllerFunc

// NewContext 创建绑定到当前应用的请求上下文。
func NewContext(r *http.Request, w http.ResponseWriter, application *Application) *Context {
	return requestcontext.NewContext(r, w, application)
}

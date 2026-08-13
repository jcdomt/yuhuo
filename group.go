package yuhuo

import (
	"net/http"

	requestcontext "github.com/jcdomt/yuhuo/context"
	"github.com/jcdomt/yuhuo/router"
)

// ApplicationGroup 表示共享路径前缀和中间件的路由分组。
type ApplicationGroup struct {
	application *Application
	routerGroup *router.RouteGroup
	handlers    []HandlerFunc
}

func newApplicationGroup(app *Application, routerGroup *router.RouteGroup) *ApplicationGroup {
	return &ApplicationGroup{application: app, routerGroup: routerGroup}
}

// Group 创建继承当前中间件的子路由组。
func (group *ApplicationGroup) Group(prefix string) *ApplicationGroup {
	return &ApplicationGroup{application: group.application, routerGroup: group.routerGroup.Group(prefix), handlers: cloneHandlers(group.handlers)}
}

// Use 添加后续路由使用的中间件。
func (group *ApplicationGroup) Use(handlers ...HandlerFunc) {
	for _, handler := range handlers {
		if handler == nil {
			panic("yuhuo: handler must not be nil")
		}
	}
	group.handlers = append(group.handlers, handlers...)
}

// Handle 注册一个带当前中间件链的路由。
func (group *ApplicationGroup) Handle(method, pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	if handler == nil {
		panic("yuhuo: handler must not be nil")
	}
	// 中间件链：组中间件 + 路由中间件 + 路由处理函数
	handlers := append(cloneHandlers(group.handlers), middlewares...)
	handlers = append(handlers, handler)

	err := group.routerGroup.Handle(method, pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := requestcontext.NewContext(r, w, group.application)
		ctx.SetHandlers(handlers)
		ctx.Next()
	}))
	if err != nil {
		panic(err)
	}
}

func (group *ApplicationGroup) GET(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	group.Handle(http.MethodGet, pattern, handler, middlewares...)
}
func (group *ApplicationGroup) POST(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	group.Handle(http.MethodPost, pattern, handler, middlewares...)
}
func (group *ApplicationGroup) PUT(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	group.Handle(http.MethodPut, pattern, handler, middlewares...)
}
func (group *ApplicationGroup) DELETE(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	group.Handle(http.MethodDelete, pattern, handler, middlewares...)
}
func (group *ApplicationGroup) PATCH(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	group.Handle(http.MethodPatch, pattern, handler, middlewares...)
}
func (group *ApplicationGroup) HEAD(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	group.Handle(http.MethodHead, pattern, handler, middlewares...)
}
func (group *ApplicationGroup) OPTIONS(pattern string, handler HandlerFunc, middlewares ...HandlerFunc) {
	group.Handle(http.MethodOptions, pattern, handler, middlewares...)
}

func cloneHandlers(handlers []HandlerFunc) []HandlerFunc {
	return append([]HandlerFunc(nil), handlers...)
}

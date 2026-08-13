package yuhuo

import (
	"net/http"

	"github.com/jcdomt/yuhuo/router"
)

type ApplicationGroup struct {
	// root APP
	application *Application

	routerGroup *router.RouteGroup
	handlers    []HandlerFunc
}

func newApplicationGroup(app *Application, routerGroup *router.RouteGroup) *ApplicationGroup {
	return &ApplicationGroup{
		application: app,
		routerGroup: routerGroup,
	}
}

func (group *ApplicationGroup) Group(prefix string) *ApplicationGroup {
	return &ApplicationGroup{
		routerGroup: group.routerGroup.Group(prefix),
		handlers:    cloneHandlers(group.handlers),
	}
}

// 中间件
func (group *ApplicationGroup) Use(handlers ...HandlerFunc) {
	for _, handler := range handlers {
		if handler == nil {
			panic("yuhuo: handler must not be nil")
		}
	}
	group.handlers = append(group.handlers, handlers...)
}

func (group *ApplicationGroup) Handle(method, pattern string, handler HandlerFunc) {
	if handler == nil {
		panic("yuhuo: handler must not be nil")
	}

	handlers := append(cloneHandlers(group.handlers), handler)
	err := group.routerGroup.Handle(method, pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := NewContext(r, w, group.application)
		ctx.handlers = handlers
		ctx.Next()
	}))
	if err != nil {
		panic(err)
	}
}

func (group *ApplicationGroup) GET(pattern string, handler HandlerFunc) {
	group.Handle(http.MethodGet, pattern, handler)
}

func (group *ApplicationGroup) POST(pattern string, handler HandlerFunc) {
	group.Handle(http.MethodPost, pattern, handler)
}

func (group *ApplicationGroup) PUT(pattern string, handler HandlerFunc) {
	group.Handle(http.MethodPut, pattern, handler)
}

func (group *ApplicationGroup) DELETE(pattern string, handler HandlerFunc) {
	group.Handle(http.MethodDelete, pattern, handler)
}

func (group *ApplicationGroup) PATCH(pattern string, handler HandlerFunc) {
	group.Handle(http.MethodPatch, pattern, handler)
}

func (group *ApplicationGroup) HEAD(pattern string, handler HandlerFunc) {
	group.Handle(http.MethodHead, pattern, handler)
}

func (group *ApplicationGroup) OPTIONS(pattern string, handler HandlerFunc) {
	group.Handle(http.MethodOptions, pattern, handler)
}

func cloneHandlers(handlers []HandlerFunc) []HandlerFunc {
	return append([]HandlerFunc(nil), handlers...)
}

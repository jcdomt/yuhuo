package mvc

import (
	"net/http"
	"strings"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// ControllerRouteGroup 描述 MVC 注册路由所需的最小能力。
type ControllerRouteGroup interface {
	GET(pattern string, handler requestcontext.HandlerFunc)
	POST(pattern string, handler requestcontext.HandlerFunc)
	PUT(pattern string, handler requestcontext.HandlerFunc)
	DELETE(pattern string, handler requestcontext.HandlerFunc)
	PATCH(pattern string, handler requestcontext.HandlerFunc)
	HEAD(pattern string, handler requestcontext.HandlerFunc)
	OPTIONS(pattern string, handler requestcontext.HandlerFunc)
}

// BaseController 为控制器提供当前请求上下文的约定字段。
type BaseController struct {
	Ctx *requestcontext.Context
}

// ControllerRouter 是控制器声明路由时使用的受限接口。
type ControllerRouter struct {
	group  ControllerRouteGroup
	prefix string
}

// Handle 按 HTTP 方法注册控制器路由。
func (router ControllerRouter) Handle(method, path string, handler requestcontext.HandlerFunc) {
	if router.group == nil {
		panic("yuhuo/mvc: controller router is not initialized")
	}
	switch method {
	case http.MethodGet:
		router.group.GET(joinControllerPath(router.prefix, path), handler)
	case http.MethodPost:
		router.group.POST(joinControllerPath(router.prefix, path), handler)
	case http.MethodPut:
		router.group.PUT(joinControllerPath(router.prefix, path), handler)
	case http.MethodDelete:
		router.group.DELETE(joinControllerPath(router.prefix, path), handler)
	case http.MethodPatch:
		router.group.PATCH(joinControllerPath(router.prefix, path), handler)
	case http.MethodHead:
		router.group.HEAD(joinControllerPath(router.prefix, path), handler)
	case http.MethodOptions:
		router.group.OPTIONS(joinControllerPath(router.prefix, path), handler)
	default:
		panic("yuhuo/mvc: unsupported controller method: " + method)
	}
}

// Group 创建一个新的路由组，并在该组中注册控制器路由。
func (router ControllerRouter) Group(path string, fn func(r ControllerRouter)) {
	if router.group == nil {
		panic("yuhuo/mvc: controller router is not initialized")
	}
	if fn == nil {
		panic("yuhuo/mvc: controller group callback must not be nil")
	}
	fn(ControllerRouter{group: router.group, prefix: joinControllerPath(router.prefix, path)})
}

// joinControllerPath 合并控制器路由前缀和相对路径。
func joinControllerPath(prefix, path string) string {
	if path == "" || path == "/" {
		if prefix == "" {
			return "/"
		}
		return prefix
	}
	if !strings.HasPrefix(path, "/") {
		panic("yuhuo/mvc: controller route must begin with '/'")
	}
	if prefix == "" {
		return path
	}
	return strings.TrimSuffix(prefix, "/") + path
}

// 将 ControllerFunc 转换为 requestcontext.HandlerFunc，以便在路由中使用。
func controllerFuncToHandlerFunc(handler requestcontext.ControllerFunc) requestcontext.HandlerFunc {
	if handler == nil {
		panic("yuhuo/mvc: controller function must not be nil")
	}

	return func(ctx *requestcontext.Context) {
		ctx.JSON(handler())
	}
}

// GET 注册 GET 路由。
func (router ControllerRouter) GET(path string, handler requestcontext.ControllerFunc) {
	router.Handle(http.MethodGet, path, controllerFuncToHandlerFunc(handler))
}

// POST 注册 POST 路由。
func (router ControllerRouter) POST(path string, handler requestcontext.ControllerFunc) {
	router.Handle(http.MethodPost, path, controllerFuncToHandlerFunc(handler))
}

// PUT 注册 PUT 路由。
func (router ControllerRouter) PUT(path string, handler requestcontext.ControllerFunc) {
	router.Handle(http.MethodPut, path, controllerFuncToHandlerFunc(handler))
}

// DELETE 注册 DELETE 路由。
func (router ControllerRouter) DELETE(path string, handler requestcontext.ControllerFunc) {
	router.Handle(http.MethodDelete, path, controllerFuncToHandlerFunc(handler))
}

// PATCH 注册 PATCH 路由。
func (router ControllerRouter) PATCH(path string, handler requestcontext.ControllerFunc) {
	router.Handle(http.MethodPatch, path, controllerFuncToHandlerFunc(handler))
}

// HEAD 注册 HEAD 路由。
func (router ControllerRouter) HEAD(path string, handler requestcontext.ControllerFunc) {
	router.Handle(http.MethodHead, path, controllerFuncToHandlerFunc(handler))
}

// OPTIONS 注册 OPTIONS 路由。
func (router ControllerRouter) OPTIONS(path string, handler requestcontext.ControllerFunc) {
	router.Handle(http.MethodOptions, path, controllerFuncToHandlerFunc(handler))
}

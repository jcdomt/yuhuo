package mvc

import (
	"net/http"

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
	group ControllerRouteGroup
}

// Handle 按 HTTP 方法注册控制器路由。
func (router ControllerRouter) Handle(method, path string, handler requestcontext.HandlerFunc) {
	if router.group == nil {
		panic("yuhuo/mvc: controller router is not initialized")
	}
	switch method {
	case http.MethodGet:
		router.group.GET(path, handler)
	case http.MethodPost:
		router.group.POST(path, handler)
	case http.MethodPut:
		router.group.PUT(path, handler)
	case http.MethodDelete:
		router.group.DELETE(path, handler)
	case http.MethodPatch:
		router.group.PATCH(path, handler)
	case http.MethodHead:
		router.group.HEAD(path, handler)
	case http.MethodOptions:
		router.group.OPTIONS(path, handler)
	default:
		panic("yuhuo/mvc: unsupported controller method: " + method)
	}
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

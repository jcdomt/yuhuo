package mvc

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// ControllerRouteGroup 描述 MVC 注册路由所需的最小能力。
type ControllerRouteGroup interface {
	GET(pattern string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc)
	POST(pattern string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc)
	PUT(pattern string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc)
	DELETE(pattern string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc)
	PATCH(pattern string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc)
	HEAD(pattern string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc)
	OPTIONS(pattern string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc)
}

// BaseController 为控制器提供当前请求上下文的约定字段。
type BaseController struct {
	Ctx *requestcontext.Context
}

// ControllerRouter 是控制器声明路由时使用的受限接口。
type ControllerRouter struct {
	group         ControllerRouteGroup
	prefix        string
	newController func() reflect.Value
}

// Handle 按 HTTP 方法注册控制器路由。
func (router ControllerRouter) Handle(method, path string, handler requestcontext.HandlerFunc, middlewares ...requestcontext.HandlerFunc) {
	if router.group == nil {
		panic("yuhuo/mvc: controller router is not initialized")
	}
	switch method {
	case http.MethodGet:
		router.group.GET(joinControllerPath(router.prefix, path), handler, middlewares...)
	case http.MethodPost:
		router.group.POST(joinControllerPath(router.prefix, path), handler, middlewares...)
	case http.MethodPut:
		router.group.PUT(joinControllerPath(router.prefix, path), handler, middlewares...)
	case http.MethodDelete:
		router.group.DELETE(joinControllerPath(router.prefix, path), handler, middlewares...)
	case http.MethodPatch:
		router.group.PATCH(joinControllerPath(router.prefix, path), handler, middlewares...)
	case http.MethodHead:
		router.group.HEAD(joinControllerPath(router.prefix, path), handler, middlewares...)
	case http.MethodOptions:
		router.group.OPTIONS(joinControllerPath(router.prefix, path), handler, middlewares...)
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
	fn(ControllerRouter{group: router.group, prefix: joinControllerPath(router.prefix, path), newController: router.newController})
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

// controllerFuncToHandlerFunc 将控制器方法转换为请求处理函数。
// 具名控制器方法会在每次请求时创建独立实例并注入 Context，匿名函数则保持直接调用。
func controllerFuncToHandlerFunc(handler requestcontext.ControllerFunc, newController func() reflect.Value) requestcontext.HandlerFunc {
	if handler == nil {
		panic("yuhuo/mvc: controller function must not be nil")
	}

	methodName, isControllerMethod := controllerMethodName(handler)
	if !isControllerMethod || newController == nil {
		return func(ctx *requestcontext.Context) {
			ctx.JSON(handler())
		}
	}

	return func(ctx *requestcontext.Context) {
		controller := newController()
		injectControllerContext(controller, ctx)

		method := controller.MethodByName(methodName)
		if !method.IsValid() {
			panic("yuhuo/mvc: controller method not found: " + methodName)
		}

		results := method.Call(nil)
		if len(results) != 1 {
			panic("yuhuo/mvc: controller method must return exactly one value: " + methodName)
		}
		ctx.JSON(results[0].Interface())
	}
}

// controllerMethodName 从方法值中提取控制器方法名称。
// Go 为方法值生成的运行时名称会以 "-fm" 结尾，匿名函数不具备此特征。
func controllerMethodName(handler requestcontext.ControllerFunc) (string, bool) {
	function := runtime.FuncForPC(reflect.ValueOf(handler).Pointer())
	if function == nil {
		return "", false
	}
	name := function.Name()
	if !strings.HasSuffix(name, "-fm") {
		return "", false
	}

	methodName := strings.TrimSuffix(name[strings.LastIndex(name, ".")+1:], "-fm")
	return methodName, methodName != ""
}

// injectControllerContext 将当前请求上下文注入控制器嵌入的 BaseController。
func injectControllerContext(controller reflect.Value, ctx *requestcontext.Context) {
	if controller.Kind() != reflect.Ptr || controller.IsNil() {
		return
	}
	setBaseControllerContext(controller.Elem(), ctx)
}

// setBaseControllerContext 遍历匿名嵌入字段，支持多层嵌入 BaseController。
func setBaseControllerContext(value reflect.Value, ctx *requestcontext.Context) bool {
	if value.Kind() != reflect.Struct {
		return false
	}

	baseControllerType := reflect.TypeOf(BaseController{})
	for index := 0; index < value.NumField(); index++ {
		field := value.Field(index)
		fieldType := value.Type().Field(index)
		if fieldType.Type == baseControllerType && field.CanSet() {
			field.FieldByName("Ctx").Set(reflect.ValueOf(ctx))
			return true
		}
		if !fieldType.Anonymous {
			continue
		}
		if field.Kind() == reflect.Ptr {
			if field.Type().Elem().Kind() != reflect.Struct || !field.CanSet() {
				continue
			}
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		if setBaseControllerContext(field, ctx) {
			return true
		}
	}
	return false
}

// GET 注册 GET 路由。
func (router ControllerRouter) GET(path string, handler requestcontext.ControllerFunc, middlewares ...requestcontext.HandlerFunc) {
	router.Handle(http.MethodGet, path, controllerFuncToHandlerFunc(handler, router.newController), middlewares...)
}

// POST 注册 POST 路由。
func (router ControllerRouter) POST(path string, handler requestcontext.ControllerFunc, middlewares ...requestcontext.HandlerFunc) {
	router.Handle(http.MethodPost, path, controllerFuncToHandlerFunc(handler, router.newController), middlewares...)
}

// PUT 注册 PUT 路由。
func (router ControllerRouter) PUT(path string, handler requestcontext.ControllerFunc, middlewares ...requestcontext.HandlerFunc) {
	router.Handle(http.MethodPut, path, controllerFuncToHandlerFunc(handler, router.newController), middlewares...)
}

// DELETE 注册 DELETE 路由。
func (router ControllerRouter) DELETE(path string, handler requestcontext.ControllerFunc, middlewares ...requestcontext.HandlerFunc) {
	router.Handle(http.MethodDelete, path, controllerFuncToHandlerFunc(handler, router.newController), middlewares...)
}

// PATCH 注册 PATCH 路由。
func (router ControllerRouter) PATCH(path string, handler requestcontext.ControllerFunc, middlewares ...requestcontext.HandlerFunc) {
	router.Handle(http.MethodPatch, path, controllerFuncToHandlerFunc(handler, router.newController), middlewares...)
}

// HEAD 注册 HEAD 路由。
func (router ControllerRouter) HEAD(path string, handler requestcontext.ControllerFunc, middlewares ...requestcontext.HandlerFunc) {
	router.Handle(http.MethodHead, path, controllerFuncToHandlerFunc(handler, router.newController), middlewares...)
}

// OPTIONS 注册 OPTIONS 路由。
func (router ControllerRouter) OPTIONS(path string, handler requestcontext.ControllerFunc, middlewares ...requestcontext.HandlerFunc) {
	router.Handle(http.MethodOptions, path, controllerFuncToHandlerFunc(handler, router.newController), middlewares...)
}

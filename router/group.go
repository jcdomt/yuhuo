package router

import (
	"net/http"
	"strings"
)

// Middleware 包装一个标准 HTTP 处理器
type Middleware func(next http.Handler) http.Handler

// RouteGroup 注册带有共享路径前缀和中间件链的路由
// 所有路由都存储在同一个 Router 字典树中
type RouteGroup struct {
	router      *Router
	prefix      string
	middlewares []Middleware
}

// Group	创建一个以 prefix 为根的路由组
//
// param:
//   - prefix	路由组路径前缀
// return:
//   - 路由组
func (r *Router) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		router: r,
		prefix: normalizePrefix(prefix),
	}
}

// Group	创建继承父组中间件的嵌套路由组
//
// param:
//   - prefix	嵌套路由组路径前缀
// return:
//   - 嵌套路由组
func (group *RouteGroup) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		router:      group.router,
		prefix:      joinRoutePath(group.prefix, normalizePrefix(prefix)),
		middlewares: cloneMiddlewares(group.middlewares),
	}
}

// Use	添加后续注册路由使用的中间件
//
// param:
//   - middlewares	中间件列表
func (group *RouteGroup) Use(middlewares ...Middleware) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// Handle	注册一条路由
//
// param:
//   - method	HTTP 方法
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Handle(method, pattern string, handler http.Handler) error {
	return group.HandleWithSource(method, pattern, handler, "")
}

// HandleWithSource	注册路由并附带处理器的定义位置
//
// param:
//   - method	HTTP 方法
//   - pattern	路由模式
//   - handler	处理器
//   - source	处理器定义位置
// return:
//   - 注册过程中的错误
func (group *RouteGroup) HandleWithSource(method, pattern string, handler http.Handler, source string) error {
	if handler == nil {
		return ErrHandlerMustNotBeNil
	}
	if source == "" {
		source = HandlerSource(handler)
	}
	return group.router.HandleWithSource(method, joinRoutePath(group.prefix, pattern), applyMiddlewares(handler, group.middlewares), source)
}

// Get	注册 GET 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Get(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodGet, pattern, handler)
}

// Post	注册 POST 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Post(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodPost, pattern, handler)
}

// Put	注册 PUT 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Put(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodPut, pattern, handler)
}

// Patch	注册 PATCH 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Patch(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodPatch, pattern, handler)
}

// Delete	注册 DELETE 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Delete(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodDelete, pattern, handler)
}

// Head	注册 HEAD 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Head(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodHead, pattern, handler)
}

// Options	注册 OPTIONS 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (group *RouteGroup) Options(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodOptions, pattern, handler)
}

// applyMiddlewares	将中间件链应用到处理器上
//
// param:
//   - handler	原始处理器
//   - middlewares	中间件列表
// return:
//   - 包装后的处理器
func applyMiddlewares(handler http.Handler, middlewares []Middleware) http.Handler {
	for index := len(middlewares) - 1; index >= 0; index-- {
		if middlewares[index] == nil {
			panic("router: middleware must not be nil")
		}
		handler = middlewares[index](handler)
	}
	return handler
}

// cloneMiddlewares	复制中间件切片，避免共享引用导致的副作用
//
// param:
//   - middlewares	中间件列表
// return:
//   - 复制后的中间件列表
func cloneMiddlewares(middlewares []Middleware) []Middleware {
	return append([]Middleware(nil), middlewares...)
}

// normalizePrefix	规范化路由组前缀，去掉首尾多余的斜杠
//
// param:
//   - prefix	路由组前缀
// return:
//   - 规范化后的前缀
func normalizePrefix(prefix string) string {
	if prefix == "" || prefix == "/" {
		return ""
	}
	if !strings.HasPrefix(prefix, "/") {
		panic("router: group prefix must begin with '/'")
	}
	return strings.TrimSuffix(prefix, "/")
}

// joinRoutePath	合并路由组前缀和路由模式
//
// param:
//   - prefix	路由组前缀
//   - pattern	路由模式
// return:
//   - 合并后的完整路径
func joinRoutePath(prefix, pattern string) string {
	if pattern == "" || pattern == "/" {
		if prefix == "" {
			return "/"
		}
		return prefix
	}
	if !strings.HasPrefix(pattern, "/") {
		panic("router: route pattern must begin with '/'")
	}
	return prefix + pattern
}

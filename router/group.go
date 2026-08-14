package router

import (
	"net/http"
	"strings"
)

// Middleware wraps a standard HTTP handler.
type Middleware func(next http.Handler) http.Handler

// RouteGroup registers routes with a shared path prefix and middleware chain.
// All routes are stored in the same Router radix tree.
type RouteGroup struct {
	router      *Router
	prefix      string
	middlewares []Middleware
}

// Group creates a route group rooted at prefix.
func (r *Router) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		router: r,
		prefix: normalizePrefix(prefix),
	}
}

// Group creates a nested route group and inherits the parent's middlewares.
func (group *RouteGroup) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		router:      group.router,
		prefix:      joinRoutePath(group.prefix, normalizePrefix(prefix)),
		middlewares: cloneMiddlewares(group.middlewares),
	}
}

// Use adds middlewares for subsequently registered routes in this group.
func (group *RouteGroup) Use(middlewares ...Middleware) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (group *RouteGroup) Handle(method, pattern string, handler http.Handler) error {
	return group.HandleWithSource(method, pattern, handler, "")
}

// HandleWithSource 注册路由并附带处理器的定义位置。
func (group *RouteGroup) HandleWithSource(method, pattern string, handler http.Handler, source string) error {
	if handler == nil {
		return ErrHandlerMustNotBeNil
	}
	if source == "" {
		source = HandlerSource(handler)
	}
	return group.router.HandleWithSource(method, joinRoutePath(group.prefix, pattern), applyMiddlewares(handler, group.middlewares), source)
}

func (group *RouteGroup) Get(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodGet, pattern, handler)
}

func (group *RouteGroup) Post(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodPost, pattern, handler)
}

func (group *RouteGroup) Put(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodPut, pattern, handler)
}

func (group *RouteGroup) Patch(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodPatch, pattern, handler)
}

func (group *RouteGroup) Delete(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodDelete, pattern, handler)
}

func (group *RouteGroup) Head(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodHead, pattern, handler)
}

func (group *RouteGroup) Options(pattern string, handler http.Handler) error {
	return group.Handle(http.MethodOptions, pattern, handler)
}

func applyMiddlewares(handler http.Handler, middlewares []Middleware) http.Handler {
	for index := len(middlewares) - 1; index >= 0; index-- {
		if middlewares[index] == nil {
			panic("router: middleware must not be nil")
		}
		handler = middlewares[index](handler)
	}
	return handler
}

func cloneMiddlewares(middlewares []Middleware) []Middleware {
	return append([]Middleware(nil), middlewares...)
}

func normalizePrefix(prefix string) string {
	if prefix == "" || prefix == "/" {
		return ""
	}
	if !strings.HasPrefix(prefix, "/") {
		panic("router: group prefix must begin with '/'")
	}
	return strings.TrimSuffix(prefix, "/")
}

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

package router

import (
	"fmt"
	"net/http"
	"strings"
)

// Handle	注册一条路由
//
// param:
//   - method	HTTP 方法
//   - pattern	路由模式
//   - handler	处理器
// return:
//   - 注册过程中的错误
func (r *Router) Handle(method, pattern string, handler http.Handler) error {
	return r.HandleWithSource(method, pattern, handler, "")
}

// HandleWithSource	注册路由，并在日志中附带处理器的定义位置
//
// param:
//   - method	HTTP 方法
//   - pattern	路由模式
//   - handler	处理器
//   - source	处理器定义位置
// return:
//   - 注册过程中的错误
func (r *Router) HandleWithSource(method, pattern string, handler http.Handler, source string) error {
	if handler == nil {
		return ErrHandlerMustNotBeNil
	}
	if source == "" {
		source = HandlerSource(handler)
	}

	method = strings.ToUpper(method)
	if method == "" {
		return ErrMethodMustNotBeEmpty
	}

	parts := parsePattern(pattern)

	r.mu.Lock()
	defer r.mu.Unlock()

	// 注入树根节点
	root := r.trees[method]
	if root == nil {
		root = &node{kind: staticNode}
		r.trees[method] = root
	}

	current := root
	for _, part := range parts {
		switch part.kind {
		case staticNode:
			current = insertStatic(current, part.value)
		case paramNode, catchAllNode:
			current = insertWildcard(current, part.kind, part.value)
		}
	}

	if current.handler != nil {
		return fmt.Errorf("%w: %s %s", ErrRouteAlreadyRegistered, method, pattern)
	}
	current.handler = handler
	current.source = source

	if r.logger != nil {
		if source != "" {
			r.logger.Debug("注册路由：", method, " ", pattern, "\t→\t", source)
		} else {
			r.logger.Debug("注册路由：", method, " ", pattern)
		}
	}

	return nil
}

// Get	注册 GET 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
func (r *Router) Get(pattern string, handler http.Handler) {
	r.Handle(http.MethodGet, pattern, handler)
}

// Post	注册 POST 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
func (r *Router) Post(pattern string, handler http.Handler) {
	r.Handle(http.MethodPost, pattern, handler)
}

// Put	注册 PUT 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
func (r *Router) Put(pattern string, handler http.Handler) {
	r.Handle(http.MethodPut, pattern, handler)
}

// Patch	注册 PATCH 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
func (r *Router) Patch(pattern string, handler http.Handler) {
	r.Handle(http.MethodPatch, pattern, handler)
}

// Delete	注册 DELETE 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
func (r *Router) Delete(pattern string, handler http.Handler) {
	r.Handle(http.MethodDelete, pattern, handler)
}

// Head	注册 HEAD 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
func (r *Router) Head(pattern string, handler http.Handler) {
	r.Handle(http.MethodHead, pattern, handler)
}

// Options	注册 OPTIONS 路由
//
// param:
//   - pattern	路由模式
//   - handler	处理器
func (r *Router) Options(pattern string, handler http.Handler) {
	r.Handle(http.MethodOptions, pattern, handler)
}

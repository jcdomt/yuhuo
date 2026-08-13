package router

import (
	"fmt"
	"net/http"
	"strings"
)

func (r *Router) Handle(method, pattern string, handler http.Handler) error {
	if handler == nil {
		return ErrHandlerMustNotBeNil
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
		return fmt.Errorf("router: route already registered: %s %s", method, pattern)
	}
	current.handler = handler

	return nil
}

func (r *Router) Get(pattern string, handler http.Handler) {
	r.Handle(http.MethodGet, pattern, handler)
}

func (r *Router) Post(pattern string, handler http.Handler) {
	r.Handle(http.MethodPost, pattern, handler)
}

func (r *Router) Put(pattern string, handler http.Handler) {
	r.Handle(http.MethodPut, pattern, handler)
}

func (r *Router) Patch(pattern string, handler http.Handler) {
	r.Handle(http.MethodPatch, pattern, handler)
}

func (r *Router) Delete(pattern string, handler http.Handler) {
	r.Handle(http.MethodDelete, pattern, handler)
}

func (r *Router) Head(pattern string, handler http.Handler) {
	r.Handle(http.MethodHead, pattern, handler)
}

func (r *Router) Options(pattern string, handler http.Handler) {
	r.Handle(http.MethodOptions, pattern, handler)
}

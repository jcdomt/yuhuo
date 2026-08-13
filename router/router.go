package router

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type nodeKind uint8

const (
	staticNode nodeKind = iota
	paramNode
	catchAllNode
)

var (
	ErrHandlerMustNotBeNil  = errors.New("router: handler must not be nil")
	ErrMethodMustNotBeEmpty = errors.New("router: method must not be empty")
)

// 压缩字典树的节点结构体
type node struct {
	kind      nodeKind
	path      string
	paramName string
	handler   http.Handler
	children  []*node
}

// 使用压缩字典树
type Router struct {
	mu    sync.RWMutex
	trees map[string]*node
}

type routePart struct {
	kind  nodeKind
	value string
}

type paramsKey struct{}

func GetDefaultRouter() *Router {
	return &Router{
		trees: make(map[string]*node),
	}
}

// Param 用于获取路由参数的值。它从请求的上下文中提取参数映射，并返回指定参数名称的值。
func Param(req *http.Request, name string) string {
	params, _ := req.Context().Value(paramsKey{}).(map[string]string)
	return params[name]
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if match, ok := matchRoute(r.trees[req.Method], req.URL.Path, nil); ok {
		if len(match.params) > 0 {
			req = req.WithContext(context.WithValue(req.Context(), paramsKey{}, match.params))
		}
		match.handler.ServeHTTP(w, req)
		return
	}

	allowed := r.allowedMethods(req.URL.Path)
	if len(allowed) > 0 {
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, req)
}

func (r *Router) allowedMethods(path string) []string {
	methods := make([]string, 0)
	for method, root := range r.trees {
		if _, ok := matchRoute(root, path, nil); ok {
			methods = append(methods, method)
		}
	}
	sort.Strings(methods)
	return methods
}

func parsePattern(pattern string) []routePart {
	if pattern == "/" {
		return []routePart{{kind: staticNode, value: "/"}}
	}
	if !strings.HasPrefix(pattern, "/") {
		panic("router: pattern must begin with '/': " + pattern)
	}

	segments := strings.Split(strings.TrimPrefix(pattern, "/"), "/")
	parts := make([]routePart, 0, len(segments))
	staticPath := ""

	flushStatic := func() {
		if staticPath != "" {
			parts = append(parts, routePart{kind: staticNode, value: staticPath})
			staticPath = ""
		}
	}

	for index, segment := range segments {
		if segment == "" {
			panic("router: empty path segment in pattern: " + pattern)
		}

		if segment[0] == ':' || segment[0] == '*' {
			if len(segment) == 1 {
				panic("router: parameter name must not be empty: " + pattern)
			}
			if strings.ContainsAny(segment[1:], ":*") {
				panic("router: parameter must occupy its entire path segment: " + pattern)
			}
			if segment[0] == '*' && index != len(segments)-1 {
				panic("router: catch-all parameter must be the last path segment: " + pattern)
			}

			// 动态部分前的斜杠属于前面的静态片段，因此参数匹配仅接收该部分的值。
			staticPath += "/"
			flushStatic()
			kind := paramNode
			if segment[0] == '*' {
				kind = catchAllNode
			}
			parts = append(parts, routePart{kind: kind, value: segment[1:]})
			continue
		}

		staticPath += "/" + segment
	}
	flushStatic()
	return parts
}

func insertStatic(parent *node, path string) *node {
	for path != "" {
		child := findStaticChild(parent, path[0])
		if child == nil {
			child = &node{kind: staticNode, path: path}
			parent.children = append(parent.children, child)
			return child
		}

		common := commonPrefixLength(child.path, path)
		if common == len(child.path) {
			parent = child
			path = path[common:]
			continue
		}

		split := &node{
			kind:     staticNode,
			path:     child.path[:common],
			children: []*node{child},
		}
		child.path = child.path[common:]
		replaceChild(parent, child, split)

		if common == len(path) {
			return split
		}

		newChild := &node{kind: staticNode, path: path[common:]}
		split.children = append(split.children, newChild)
		return newChild
	}

	return parent
}

func insertWildcard(parent *node, kind nodeKind, name string) *node {
	for _, child := range parent.children {
		if child.kind != kind {
			continue
		}
		if child.paramName != name {
			panic("router: conflicting parameter names at the same path level")
		}
		return child
	}

	child := &node{kind: kind, paramName: name}
	parent.children = append(parent.children, child)
	return child
}

func findStaticChild(parent *node, firstByte byte) *node {
	for _, child := range parent.children {
		if child.kind == staticNode && child.path[0] == firstByte {
			return child
		}
	}
	return nil
}

func replaceChild(parent, oldChild, newChild *node) {
	for index, child := range parent.children {
		if child == oldChild {
			parent.children[index] = newChild
			return
		}
	}
}

func commonPrefixLength(left, right string) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}

	index := 0
	for index < limit && left[index] == right[index] {
		index++
	}
	return index
}

type matchResult struct {
	handler http.Handler
	params  map[string]string
}

func matchRoute(root *node, path string, params map[string]string) (matchResult, bool) {
	if root == nil {
		return matchResult{}, false
	}

	return matchNode(root, path, params)
}

func matchNode(current *node, path string, params map[string]string) (matchResult, bool) {
	if path == "" && current.handler != nil {
		return matchResult{handler: current.handler, params: params}, true
	}

	for _, child := range current.children {
		if child.kind != staticNode || !strings.HasPrefix(path, child.path) {
			continue
		}
		if match, ok := matchNode(child, path[len(child.path):], params); ok {
			return match, true
		}
	}

	for _, child := range current.children {
		if child.kind != paramNode {
			continue
		}
		value, rest, ok := nextSegment(path)
		if !ok {
			continue
		}
		if match, ok := matchNode(child, rest, withParam(params, child.paramName, value)); ok {
			return match, true
		}
	}

	for _, child := range current.children {
		if child.kind != catchAllNode || path == "" {
			continue
		}
		if child.handler != nil {
			return matchResult{
				handler: child.handler,
				params:  withParam(params, child.paramName, strings.TrimPrefix(path, "/")),
			}, true
		}
	}

	return matchResult{}, false
}

func nextSegment(path string) (value, rest string, ok bool) {
	if path == "" {
		return "", "", false
	}
	if index := strings.IndexByte(path, '/'); index >= 0 {
		if index == 0 {
			return "", "", false
		}
		return path[:index], path[index:], true
	}
	return path, "", true
}

func withParam(params map[string]string, name, value string) map[string]string {
	result := make(map[string]string, len(params)+1)
	for key, existingValue := range params {
		result[key] = existingValue
	}
	result[name] = value
	return result
}

package router

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// nodeKind 表示路由树节点的类型
type nodeKind uint8

const (
	// staticNode 静态路径节点
	staticNode nodeKind = iota
	// paramNode 参数节点，如 /users/:id
	paramNode
	// catchAllNode 通配节点，如 /assets/*path
	catchAllNode
)

var (
	// ErrHandlerMustNotBeNil 处理器不能为 nil
	ErrHandlerMustNotBeNil = errors.New("router: handler must not be nil")
	// ErrMethodMustNotBeEmpty HTTP 方法不能为空
	ErrMethodMustNotBeEmpty = errors.New("router: method must not be empty")
	// ErrRouteAlreadyRegistered 路由重复注册
	ErrRouteAlreadyRegistered = errors.New("router: route already registered")
)

// node 是压缩字典树的节点结构体
type node struct {
	kind      nodeKind
	path      string
	paramName string
	handler   http.Handler
	source    string
	children  []*node
}

// Logger 是路由器注册路由时所需的最小日志能力
// 接口定义在 router 包内，避免反向依赖根包或 context 包造成循环导入
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
}

// noopLogger 在未注入日志器时静默丢弃所有日志
type noopLogger struct{}

func (noopLogger) Debug(args ...interface{}) {}
func (noopLogger) Info(args ...interface{})  {}

// Router 使用压缩字典树存储路由
type Router struct {
	mu     sync.RWMutex
	trees  map[string]*node
	logger Logger
}

// routePart 表示解析后的路由片段
type routePart struct {
	kind  nodeKind
	value string
}

// paramsKey 是路由参数在请求上下文中的键
type paramsKey struct{}

// GetDefaultRouter	返回一个默认的路由器实例
//
// return:
//   - 路由器实例
func GetDefaultRouter() *Router {
	return &Router{
		trees:  make(map[string]*node),
		logger: noopLogger{},
	}
}

// SetLogger	注入路由注册使用的日志器
//
// param:
//   - logger	日志器，为 nil 时忽略
func (r *Router) SetLogger(logger Logger) {
	if logger == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.logger = logger
}

// Param	用于获取路由参数的值，它从请求的上下文中提取参数映射，并返回指定参数名称的值
//
// param:
//   - req	HTTP 请求
//   - name	路由参数名
// return:
//   - 路由参数值
func Param(req *http.Request, name string) string {
	params, _ := req.Context().Value(paramsKey{}).(map[string]string)
	return params[name]
}

// ServeHTTP	处理一次 HTTP 请求，匹配路由并分发到对应处理器
//
// param:
//   - w	响应写入器
//   - req	HTTP 请求
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if match, ok := matchRoute(r.trees[req.Method], req.URL.Path, nil); ok {
		if len(match.params) > 0 {
			req = req.WithContext(context.WithValue(req.Context(), paramsKey{}, match.params))
		}
		r.logRequest(req.Method, req.URL.Path, match.source)
		match.handler.ServeHTTP(w, req)
		return
	}

	allowed := r.allowedMethods(req.URL.Path)
	if len(allowed) > 0 {
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		r.logUnmatched(req.Method, req.URL.Path, "method not allowed")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	r.logUnmatched(req.Method, req.URL.Path, "not found")
	http.NotFound(w, req)
}

// logRequest	记录一次匹配到的请求日志
//
// param:
//   - method	HTTP 方法
//   - path	请求路径
//   - source	处理器定义位置
func (r *Router) logRequest(method, path, source string) {
	if r.logger == nil {
		return
	}
	if source != "" {
		r.logger.Info("收到请求：", method, " ", path, " → ", source)
		return
	}
	r.logger.Info("收到请求：", method, " ", path)
}

// logUnmatched	记录一次未匹配的请求日志
//
// param:
//   - method	HTTP 方法
//   - path	请求路径
//   - reason	未匹配原因
func (r *Router) logUnmatched(method, path, reason string) {
	if r.logger == nil {
		return
	}
	r.logger.Debug("请求未匹配：", method, " ", path, " (", reason, ")")
}

// allowedMethods	返回指定路径在其他方法下可用的方法列表
//
// param:
//   - path	请求路径
// return:
//   - 允许的方法列表
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

// parsePattern	将路由模式解析为路由片段列表
//
// param:
//   - pattern	路由模式
// return:
//   - 路由片段列表
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

			// 动态部分前的斜杠属于前面的静态片段，因此参数匹配仅接收该部分的值
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

// insertStatic	将静态路径插入父节点的子树中，必要时进行节点分裂
//
// param:
//   - parent	父节点
//   - path	静态路径
// return:
//   - 插入后的目标节点
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

// insertWildcard	将参数或通配节点插入父节点的子树中
//
// param:
//   - parent	父节点
//   - kind	节点类型（paramNode 或 catchAllNode）
//   - name	参数名
// return:
//   - 插入后的目标节点
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

// findStaticChild	在父节点中查找以指定首字节开头的静态子节点
//
// param:
//   - parent	父节点
//   - firstByte	首字节
// return:
//   - 匹配的子节点，不存在时返回 nil
func findStaticChild(parent *node, firstByte byte) *node {
	for _, child := range parent.children {
		if child.kind == staticNode && child.path[0] == firstByte {
			return child
		}
	}
	return nil
}

// replaceChild	用新节点替换父节点下的指定子节点
//
// param:
//   - parent	父节点
//   - oldChild	旧子节点
//   - newChild	新子节点
func replaceChild(parent, oldChild, newChild *node) {
	for index, child := range parent.children {
		if child == oldChild {
			parent.children[index] = newChild
			return
		}
	}
}

// commonPrefixLength	返回两个字符串的公共前缀长度
//
// param:
//   - left	左侧字符串
//   - right	右侧字符串
// return:
//   - 公共前缀长度
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

// matchResult 是一次路由匹配的结果
type matchResult struct {
	handler http.Handler
	params  map[string]string
	source  string
}

// matchRoute	从树的根节点开始匹配路径
//
// param:
//   - root	根节点
//   - path	请求路径
//   - params	已累积的路由参数
// return:
//   - 匹配结果及是否匹配成功
func matchRoute(root *node, path string, params map[string]string) (matchResult, bool) {
	if root == nil {
		return matchResult{}, false
	}

	return matchNode(root, path, params)
}

// matchNode	递归匹配当前节点及其子节点
//
// param:
//   - current	当前节点
//   - path	剩余路径
//   - params	已累积的路由参数
// return:
//   - 匹配结果及是否匹配成功
func matchNode(current *node, path string, params map[string]string) (matchResult, bool) {
	if path == "" && current.handler != nil {
		return matchResult{handler: current.handler, params: params, source: current.source}, true
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
				source:  child.source,
			}, true
		}
	}

	return matchResult{}, false
}

// nextSegment	从路径中提取下一个路径段
//
// param:
//   - path	请求路径
// return:
//   - 路径段值、剩余路径及是否成功提取
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

// withParam	在参数映射中追加一个路由参数，返回新的映射
//
// param:
//   - params	原参数映射
//   - name	参数名
//   - value	参数值
// return:
//   - 追加后的参数映射
func withParam(params map[string]string, name, value string) map[string]string {
	result := make(map[string]string, len(params)+1)
	for key, existingValue := range params {
		result[key] = existingValue
	}
	result[name] = value
	return result
}

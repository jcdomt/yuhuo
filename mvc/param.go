// 控制器函数的参数注入
package mvc

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// paramResolver 在请求到达时从上下文中解析出一个控制器函数参数
type paramResolver func(ctx *requestcontext.Context) (reflect.Value, error)

// contextPtrType 是 *requestcontext.Context 的反射类型，用于识别上下文参数
var contextPtrType = reflect.TypeOf((*requestcontext.Context)(nil))

// buildParamResolvers	注册时为控制器函数的每个输入参数构建解析器
// 参数注入规则按类型分派：
//  1. *requestcontext.Context 直接注入请求上下文
//  2. 结构体（或结构体指针）绑定请求数据，JSON 请求体走 BindJSON，否则走 BindQuery
//  3. 基本类型（string/bool/int/uint/float）注入路由路径参数（如 /users/:id）：
//     优先按参数名与路径参数名匹配（参数名来自源码解析，见 controllerParamNames），
//     无法按名匹配时按声明顺序消费剩余路径参数
//
// 不支持的参数类型或路径参数不足时在注册时 panic，保证错误尽早暴露
//
// param:
//   - handlerType	控制器函数类型
//   - pattern	完整路由路径
//   - paramNames	参数名列表，为 nil 或数量不符时全部按顺序注入
//
// return:
//   - 与参数一一对应的解析器列表
func buildParamResolvers(handlerType reflect.Type, pattern string, paramNames []string) []paramResolver {
	numIn := handlerType.NumIn()
	if numIn == 0 {
		return nil
	}
	if len(paramNames) != numIn {
		paramNames = nil
	}

	// 解析 /user/:id 形式的路径参数名，并获取每个参数的名称
	pathParams := pathParamNames(pattern)
	usedPath := make([]bool, len(pathParams))
	resolvers := make([]paramResolver, numIn)

	// 先处理上下文与结构体参数，收集需要匹配路径参数的基本类型参数
	scalars := make([]int, 0, numIn)
	for i := 0; i < numIn; i++ {
		paramType := handlerType.In(i)
		switch {
		// MVC 函数请求 Ctx 上下文参数注入
		case paramType == contextPtrType:
			resolvers[i] = func(ctx *requestcontext.Context) (reflect.Value, error) {
				return reflect.ValueOf(ctx), nil
			}
		// 类型为结构体或结构体指针时，按请求体 JSON 或查询参数绑定
		case isStructParam(paramType):
			resolvers[i] = newStructResolver(paramType)
		// 基本类型参数注入路径参数，先按名匹配，后按顺序
		case isScalarKind(paramType.Kind()):
			scalars = append(scalars, i)
		default:
			panic(fmt.Sprintf("yuhuo/mvc: unsupported controller parameter type %s (parameter %d)", paramType, i))
		}
	}

	// 第一遍：按参数名匹配路径参数，如 (id int) 匹配 /users/:id
	unmatched := make([]int, 0, len(scalars))
	for _, i := range scalars {
		name := ""
		if paramNames != nil {
			name = paramNames[i]
		}
		matched := false
		if name != "" {
			for j, pathParam := range pathParams {
				if !usedPath[j] && pathParam == name {
					resolvers[i] = newPathParamResolver(pathParam, handlerType.In(i))
					usedPath[j] = true
					matched = true
					break
				}
			}
		}
		if !matched {
			unmatched = append(unmatched, i)
		}
	}

	// 第二遍：未按名匹配的参数按声明顺序消费剩余路径参数
	next := 0
	for _, i := range unmatched {
		for next < len(pathParams) && usedPath[next] {
			next++
		}
		if next >= len(pathParams) {
			panic(fmt.Sprintf("yuhuo/mvc: controller parameter %d (%s) has no matching path parameter in pattern %q", i, handlerType.In(i), pattern))
		}
		// 允许该参数按顺序取得路径参数值，并转换为目标类型
		resolvers[i] = newPathParamResolver(pathParams[next], handlerType.In(i))
		usedPath[next] = true
	}
	return resolvers
}

// pathParamNames	按声明顺序提取路由路径中的参数名（:id 与 *path 两种形式）
//
// param:
//   - pattern	完整路由路径
//
// return:
//   - 路径参数名列表
func pathParamNames(pattern string) []string {
	var names []string
	for _, segment := range strings.Split(pattern, "/") {
		if len(segment) > 1 && (segment[0] == ':' || segment[0] == '*') {
			names = append(names, segment[1:])
		}
	}
	return names
}

// isStructParam	判断参数是否为结构体或结构体指针
func isStructParam(paramType reflect.Type) bool {
	if paramType.Kind() == reflect.Ptr {
		return paramType.Elem().Kind() == reflect.Struct
	}
	return paramType.Kind() == reflect.Struct
}

// isScalarKind	判断参数是否为可从字符串转换的基本类型
func isScalarKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// newStructResolver	为结构体参数构建解析器
// 请求体为 JSON 时解码 Body，否则绑定查询参数
//
// param:
//   - paramType	结构体或结构体指针类型
//
// return:
//   - 参数解析器
func newStructResolver(paramType reflect.Type) paramResolver {
	isPtr := paramType.Kind() == reflect.Ptr
	structType := paramType
	if isPtr {
		structType = paramType.Elem()
	}

	return func(ctx *requestcontext.Context) (reflect.Value, error) {
		obj := reflect.New(structType)

		var err error
		if shouldBindJSON(ctx) {
			err = ctx.BindJSON(obj.Interface())
		} else {
			err = ctx.BindQuery(obj.Interface())
		}
		if err != nil {
			return reflect.Value{}, err
		}

		if isPtr {
			return obj, nil
		}
		return obj.Elem(), nil
	}
}

// shouldBindJSON	判断当前请求是否应按 JSON 请求体绑定
func shouldBindJSON(ctx *requestcontext.Context) bool {
	contentType := ctx.Request().Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json") && len(bytes.TrimSpace(ctx.Body())) > 0
}

// newPathParamResolver	为基本类型参数构建解析器，从指定路径参数取值并转换类型
//
// param:
//   - name	路径参数名
//   - paramType	参数类型
//
// return:
//   - 参数解析器
func newPathParamResolver(name string, paramType reflect.Type) paramResolver {
	return func(ctx *requestcontext.Context) (reflect.Value, error) {
		raw := ctx.Param(name)
		if raw == "" {
			return reflect.Value{}, fmt.Errorf("path parameter %q is required", name)
		}
		return convertScalar(raw, paramType)
	}
}

// convertScalar	将路径参数字符串转换为目标基本类型
//
// param:
//   - raw	参数字符串
//   - paramType	目标类型
//
// return:
//   - 转换后的值及转换错误
func convertScalar(raw string, paramType reflect.Type) (reflect.Value, error) {
	value := reflect.New(paramType).Elem()
	switch paramType.Kind() {
	case reflect.String:
		value.SetString(raw)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return reflect.Value{}, err
		}
		value.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, paramType.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		value.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, paramType.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		value.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, paramType.Bits())
		if err != nil {
			return reflect.Value{}, err
		}
		value.SetFloat(f)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported scalar type %s", paramType)
	}
	return value, nil
}

// resolveParams	在请求到达时逐个解析控制器函数参数
// 任一参数解析失败时以 400 响应并中止
//
// param:
//   - ctx	请求上下文
//   - resolvers	参数解析器列表
//
// return:
//   - 解析后的参数列表，失败时返回 nil
func resolveParams(ctx *requestcontext.Context, resolvers []paramResolver) []reflect.Value {
	args := make([]reflect.Value, len(resolvers))
	for i, resolve := range resolvers {
		arg, err := resolve(ctx)
		if err != nil {
			ctx.Abort()
			ctx.JSONs(http.StatusBadRequest, requestcontext.M{"error": err.Error(), "code": http.StatusBadRequest})
			return nil
		}
		args[i] = arg
	}
	return args
}

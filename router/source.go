package router

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
)

// HandlerSource	解析处理函数的定义位置，返回 "file:line"
// 非函数类型或无法解析时返回空字符串
//
// param:
//   - handler	处理函数
// return:
//   - 定义位置
func HandlerSource(handler interface{}) string {
	if handler == nil {
		return ""
	}

	value := reflect.ValueOf(handler)
	if value.Kind() != reflect.Func {
		return ""
	}

	function := runtime.FuncForPC(value.Pointer())
	if function == nil {
		return ""
	}

	file, line := function.FileLine(value.Pointer())
	return fmt.Sprintf("%s:%d", filepath.ToSlash(file), line)
}

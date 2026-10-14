package yuhuo

import (
	"errors"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"syscall"
)

// 放一些系统内部使用的中间件，避免用户重复实现

// Recovery	返回一个恢复中间件：捕获处理链中的 panic，
// 记录堆栈日志并返回 500，避免连接被直接中断
//
// return:
//   - 恢复中间件处理函数
func Recovery() HandlerFunc {
	return func(ctx *Context) {
		defer func() {
			err := recover()
			if err == nil {
				return
			}
			if err == http.ErrAbortHandler {
				panic(err)
			}
			if isBrokenPipe(err) {
				ctx.GetLogger().Error("请求中断（客户端连接已断开）：", err)
				return
			}

			ctx.GetLogger().Error("处理器 panic 已恢复：", err, "\n", string(debug.Stack()))
			if ctx.IsCommitted() {
				return
			}
			ctx.InternalServerError()
		}()
		ctx.Next()
	}
}

// isBrokenPipe	判断 panic 值是否为断开管道（客户端提前断开）错误
//
// param:
//   - err	panic 值
// return:
//   - 是否为断管错误
func isBrokenPipe(err interface{}) bool {
	opErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}
	return errors.Is(opErr.Err, syscall.EPIPE) || strings.Contains(strings.ToLower(opErr.Err.Error()), "broken pipe")
}

package mvc

import (
	"github.com/jcdomt/yuhuo/context"
	requestcontext "github.com/jcdomt/yuhuo/context"
)

// MVC 模式的推荐响应体
// 作为一个标准的响应体，建议所有的接口都返回该结构体
// 应当是其他响应模式兼容本结构体
// 总而言之，本结构体是对一次请求的最终封装。一切以本结构体为主
type Result interface {
	Execute(ctx *requestcontext.Context)
}

var (
	ContentTypeTextPlain = "text/plain"
	ContentTypeTextHTML  = "text/html"
	ContentTypeJSON      = "application/json"
)

type Response struct {
	Code int

	ContentType string
	Content     []byte

	// 错误信息，优先级最高
	// 如果 Error.IsError() 返回 true，则直接返回错误信息，忽略其他响应
	Error ErrorResponse

	// text/plain
	Text string
	// text/html
	HTML string
	// application/json
	JSON interface{}
}

func (r Response) Execute(ctx *requestcontext.Context) {
	if r.Error != nil && r.Error.IsError() {
		ctx.JSONs(r.Error.ErrorCode(), context.M{"error": r.Error.Error(), "code": r.Error.ErrorCode()})
		return
	}
	if r.ContentType != "" {
		ctx.Header("Content-Type", r.ContentType)
		switch r.ContentType {
		case ContentTypeTextPlain:
			ctx.Data([]byte(r.Text))
		case ContentTypeTextHTML:
			ctx.Data([]byte(r.HTML))
		case ContentTypeJSON:
			ctx.JSON(r.JSON)
		}
	}

	// 如果没有设置 ContentType，则默认按顺序处理
	if r.Text != "" {
		ctx.Data([]byte(r.Text))
		return
	}
	if r.HTML != "" {
		ctx.Data([]byte(r.HTML))
		return
	}
	if r.JSON != nil {
		ctx.JSON(r.JSON)
		return
	}
}

type ErrorResponse interface {
	IsError() bool
	ErrorCode() int
	Error() string
}

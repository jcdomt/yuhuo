package framework

import "github.com/jcdomt/yuhuo/mvc"

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func (r *ApiResponse) Response() mvc.Result {
	return mvc.Response{
		Code:        r.Code,
		ContentType: mvc.ContentTypeJSON,
		JSON:        r,
	}
}

func (r *ApiResponse) Error() bool {
	// 如果 Code 不在 200-299 范围内，则认为是错误
	return r.Code/100 != 2
}

func Api(code int, msg string, data interface{}) ApiResponse {
	return ApiResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func OK() ApiResponse {
	return Api(200, "ok", nil)
}

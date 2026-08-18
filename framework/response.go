package framerwork

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

func Api(code int, msg string, data interface{}) ApiResponse {
	return ApiResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

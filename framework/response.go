package framework

import "github.com/jcdomt/yuhuo/mvc"

// 一些系统自带的报错
// ApiResponse_ParamError 请求参数格式错误
var ApiResponse_ParamError = Api(400, "请求参数格式错误", nil)

// ApiResponse_OK 成功响应
var ApiResponse_OK = OK()

// ApiResponse 是统一的接口响应体
type ApiResponse struct {
	// Code 业务状态码
	Code int `json:"code"`
	// Msg 提示信息
	Msg string `json:"msg"`
	// Data 业务数据
	Data interface{} `json:"data"`
}

// Response	将响应体转换为 mvc.Result
//
// return:
//   - JSON 形式的 mvc.Result
func (r *ApiResponse) Response() mvc.Result {
	return mvc.Response{
		Code:        r.Code,
		ContentType: mvc.ContentTypeJSON,
		JSON:        r,
	}
}

// Error	判断当前响应是否为错误响应（状态码不在 200-299 范围内）
//
// return:
//   - 是否为错误响应
func (r *ApiResponse) Error() bool {
	// 如果 Code 不在 200-299 范围内，则认为是错误
	return r.Code/100 != 2
}

// Api	构造一个统一的接口响应
//
// param:
//   - code	业务状态码
//   - msg	提示信息
//   - data	业务数据
// return:
//   - 接口响应体
func Api(code int, msg string, data interface{}) ApiResponse {
	return ApiResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

// OK	构造一个成功响应
//
// return:
//   - 接口响应体
func OK() ApiResponse {
	return Api(200, "ok", nil)
}

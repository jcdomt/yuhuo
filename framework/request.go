package framework

import requestcontext "github.com/jcdomt/yuhuo"

// ReadJSON	读取请求体 JSON 并校验，返回 ApiResponse
// 这是对 JSON 请求体读取的最高封装，将大部分功能都集成了
//
// param:
//   - ctx	请求上下文
//   - result	绑定的结果对象，必须是指针类型
// return:
//   - ApiResponse，包含错误信息或成功信息
func ReadJSON(ctx requestcontext.Context, result interface{}) ApiResponse {
	if err := ctx.BindJSON(result); err != nil {
		return ApiResponse_ParamError
	}

	errs := DefaultValidator().Validate(result)

	if len(errs) != 0 {
		return Api(400, "请求参数格式错误", map[string]interface{}{
			"errors": errs,
		})
	}

	// 为请求体填充默认值
	if err := SetDefaults(result); err != nil {
		return Api(500, "默认值填充失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return ApiResponse_OK

}

// ReadQuery	读取查询参数并校验，返回 ApiResponse
//
// param:
//   - ctx	请求上下文
//   - result	绑定的结果对象，必须是指针类型
// return:
//   - ApiResponse，包含错误信息或成功信息
func ReadQuery(ctx requestcontext.Context, result interface{}) ApiResponse {
	if err := ctx.BindQuery(result); err != nil {
		return ApiResponse_ParamError
	}

	errs := DefaultValidator().Validate(result)

	if len(errs) != 0 {
		return Api(400, "请求参数格式错误", map[string]interface{}{
			"errors": errs,
		})
	}

	// 为请求体填充默认值
	if err := SetDefaults(result); err != nil {
		return Api(500, "默认值填充失败", map[string]interface{}{
			"error": err.Error(),
		})
	}

	return ApiResponse_OK
}

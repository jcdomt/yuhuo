package framework

import requestcontext "github.com/jcdomt/yuhuo"

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

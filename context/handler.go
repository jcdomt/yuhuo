package context

// HandlerFunc 表示一个请求处理函数或中间件。
type HandlerFunc func(ctx *Context)

// M 是用于构造 JSON 等数据的便捷映射类型。
type M map[string]interface{}

// ControllerFunc 表示一个无参数并返回响应数据的控制器方法。
// 例如：func (c Controller) Index() interface{}。
type ControllerFunc func() interface{}

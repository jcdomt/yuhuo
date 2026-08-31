package context

// HandlerFunc 表示一个请求处理函数或中间件
type HandlerFunc func(ctx *Context)

// M 是用于构造 JSON 等数据的便捷映射类型
type M map[string]interface{}

// ControllerFunc 表示一个无参数并返回响应数据的控制器方法
// 例如：func (c Controller) Index() interface{}
// type ControllerFunc func() interface{}

// 2026.08.31 更新
// 为了让 ConrtollerFunc 可以支持任意参数和返回值，新增 AnyControllerFunc 接口
// 由 AnyControllerFunc 接口来接受函数再由反射来调用
type ControllerFunc interface{}

// 保留原来的 ControllerFunc 类型定义为 NotAnyControllerFunc，表示不支持任意参数和返回值的控制器函数
type NotAnyControllerFunc func() interface{}

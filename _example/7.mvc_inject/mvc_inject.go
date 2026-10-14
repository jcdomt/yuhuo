package main

import (
	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/mvc"
)

// MVC 模式 - 控制器参数自动注入

type UserController struct {
	mvc.BaseController
}

// UserQuery 用户列表查询参数，字段通过 query 标签绑定查询参数
type UserQuery struct {
	Keyword string `query:"keyword"`
	Page    int    `query:"page"`
}

// CreateUserRequest 创建用户的 JSON 请求体
type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (c UserController) Router(r mvc.ControllerRouter) {
	r.GET("/users", c.List)    // 结构体参数，绑定查询参数
	r.GET("/users/:id", c.Get) // 基本类型参数，消费路径参数 :id
	r.POST("/users", c.Create) // 结构体指针参数，绑定 JSON 请求体
	r.GET("/hello", c.Hello)   // 注入请求上下文
}

// List 结构体参数自动绑定查询参数
func (c UserController) List(query UserQuery) interface{} {
	return yuhuo.M{
		"keyword": query.Keyword,
		"page":    query.Page,
	}
}

// Get 基本类型参数优先按参数名匹配路径参数（id 匹配 :id），无法匹配时按声明顺序消费，转换失败会返回 400
func (c UserController) Get(id int) yuhuo.M {
	return yuhuo.M{"id": id}
}

// Create 结构体指针参数自动绑定 JSON 请求体
func (c UserController) Create(body *CreateUserRequest) yuhuo.M {
	return yuhuo.M{
		"name": body.Name,
		"age":  body.Age,
	}
}

// Hello 注入当前请求上下文，同时 BaseController.Ctx 也会被注入
func (c UserController) Hello(ctx *yuhuo.Context) yuhuo.M {
	return yuhuo.M{
		"message": "Hello, World!",
		"path":    ctx.Request().URL.Path,
		"sameCtx": ctx == c.Ctx,
	}
}

func main() {
	app := yuhuo.New()
	app.Logger().SetLevel(yuhuo.LogLevelDebug)

	mvc.New(app.Group("/api")).Handle(UserController{})

	_ = app.Run(":8081")
}

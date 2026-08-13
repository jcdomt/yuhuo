package main

import (
	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/mvc"
)

// MVC 模式 - 控制器 示例

type Controller struct {
	mvc.BaseController
}

var middleware = func(c *yuhuo.Context) {
	c.Set("middleware", "This is a middleware")
	c.Next()
}

func (c Controller) Router(r mvc.ControllerRouter) {
	r.Group("/hello", func(r mvc.ControllerRouter) {
		r.GET("/world", c.Hello, middleware)
	})
	r.GET("/health", c.Health)
}

func (c Controller) Hello() interface{} {
	return yuhuo.M{
		"message":    "Hello, World!",
		"middleware": c.Ctx.Get("middleware"),
	}
}

func (c Controller) Health() interface{} {
	return yuhuo.M{"status": "ok"}
}

func main() {
	app := yuhuo.New()
	mvc.New(app.Group("/api")).Handle(Controller{})
	_ = app.Run(":8080")
}

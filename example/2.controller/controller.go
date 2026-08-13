package main

import (
	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/mvc"
)

// MVC 模式 - 控制器 示例

type Controller struct {
	mvc.BaseController
}

func (c Controller) Router(r mvc.ControllerRouter) {
	r.GET("/health", c.Health)
}

func (c Controller) Health() interface{} {
	return map[string]string{"status": "ok"}
}

func main() {
	app := yuhuo.New()
	mvc.New(app.Group("/api")).Handle(Controller{})
	_ = app.Run(":8080")
}

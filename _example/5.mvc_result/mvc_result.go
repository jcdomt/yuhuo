package main

import (
	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/mvc"
)

type Controller struct {
	mvc.BaseController
}

func (c Controller) Router(r mvc.ControllerRouter) {
	r.GET("/a", c.A)
	r.GET("/b", c.B)
}

func (c Controller) A() mvc.Result {
	return mvc.Response{
		JSON: yuhuo.M{"message": "Hello, world!"},
	}
}

func (c Controller) B() mvc.Result {
	return framework.Api(200, "Success", yuhuo.M{"data": "This is a response from B"})
}

func main() {
	app := yuhuo.New()

	mvc.New(app.Group("/")).Handle(Controller{})

	app.Run(":8080")
}

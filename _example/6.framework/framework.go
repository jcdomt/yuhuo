package main

import (
	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/framework"
	"github.com/jcdomt/yuhuo/mvc"
)

type Controller struct {
	mvc.BaseController
}

func (c *Controller) Router(r mvc.ControllerRouter) {
	r.GET("/hello", c.Hello)
}

func (c *Controller) Hello() string {
	return "Hello from sub_app!"
}

type UserController struct {
	mvc.BaseController
	BaseUrl string
}

func (c *UserController) Router(r mvc.ControllerRouter) {
	r.GET("/", c.GetUsers)
}

func (c *UserController) GetUsers() interface{} {
	return "Users from sub_app!"
}

func main() {
	sub_app := framework.App{
		Url: "/sub_app",
		Controllers: framework.AppControllers{
			&Controller{},
			&UserController{BaseUrl: "/users"},
		},
	}

	framework.RegisterApp(&sub_app)

	app := yuhuo.New()
	app.Logger().SetLevel(yuhuo.LogLevelDebug)

	err := framework.RunFramework(app, ":8081")

	if err != nil {
		app.Logger().Error("Failed to run the application: ", err)
	}
}

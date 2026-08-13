package mvc

// Application 是独立的 MVC 注册器，持有目标路由组但不依赖根包应用。
type Application struct {
	router ControllerRouteGroup
}

// New 基于一个路由组创建 MVC 注册器。
func New(router ControllerRouteGroup) *Application {
	if router == nil {
		panic("yuhuo/mvc: controller route group must not be nil")
	}
	return &Application{router: router}
}

// Handle 将控制器声明的路由注册到当前 MVC 路由组。
func (app *Application) Handle(controllers ...interface{}) *Application {
	for _, controller := range controllers {
		RegisterControllerWithRouteGroup(app.router, controller)
	}
	return app
}

// RegisterControllerWithRouteGroup 将单个控制器注册到指定路由组。
func RegisterControllerWithRouteGroup(routeGroup ControllerRouteGroup, controller interface{}) {
	if routeGroup == nil {
		panic("yuhuo/mvc: controller route group must not be nil")
	}
	if c, ok := controller.(interface{ Router(ControllerRouter) }); ok {
		c.Router(ControllerRouter{group: routeGroup})
	}
}

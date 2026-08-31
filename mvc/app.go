package mvc

import (
	"reflect"
)

// Application 是独立的 MVC 注册器，持有目标路由组但不依赖根包应用
type Application struct {
	router ControllerRouteGroup
}

// New	基于一个路由组创建 MVC 注册器
//
// param:
//   - router	目标路由组，需实现 ControllerRouteGroup 接口
//
// return:
//   - MVC 注册器
func New(router ControllerRouteGroup) *Application {
	if router == nil {
		panic("yuhuo/mvc: controller route group must not be nil")
	}
	return &Application{router: router}
}

// Handle	将控制器声明的路由注册到当前 MVC 路由组
//
// param:
//   - controllers	控制器列表，需实现 Router(ControllerRouter) 方法
//
// return:
//   - 当前 MVC 注册器，便于链式调用
func (app *Application) Handle(controllers ...interface{}) *Application {
	for _, controller := range controllers {
		RegisterControllerWithRouteGroup(app.router, controller)
	}
	return app
}

// RegisterControllerWithRouteGroup	将单个控制器注册到指定路由组
//
// param:
//   - routeGroup	目标路由组
//   - controller	控制器实例，需实现 Router(ControllerRouter) 方法
func RegisterControllerWithRouteGroup(routeGroup ControllerRouteGroup, controller interface{}) {
	if routeGroup == nil {
		panic("yuhuo/mvc: controller route group must not be nil")
	}
	factory := newControllerFactory(controller)
	if c, ok := controller.(interface{ Router(ControllerRouter) }); ok {
		// 判断控制器是否申明了 BaseUrl 参数
		// 如果控制器设置了 BaseUrl，则将其作为前缀注册路由
		baseUrl := getControllerBaseUrl(c)
		c.Router(ControllerRouter{group: routeGroup, newController: factory, prefix: baseUrl})
	}
}

// getControllerBaseUrl	获取控制器的基础前缀
//
// param:
//   - controller	控制器实例
//
// return:
//   - 控制器的基础前缀，如果控制器未设置 BaseUrl，则返回空字符串
func getControllerBaseUrl(controller interface{}) string {
	// 我们是设置了 BaseUrl 的优先级的
	// 优先使用 controller.BaseController.BaseUrl，如果没有设置，则使用 controller.BaseUrl
	// 如果都没有设置，则返回空字符串
	reflectValue := reflect.ValueOf(controller)
	if reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	baseControllerField := reflectValue.FieldByName("BaseController")
	if baseControllerField.IsValid() {
		baseControllerBaseUrlField := baseControllerField.FieldByName("BaseUrl")
		if baseControllerBaseUrlField.IsValid() && baseControllerBaseUrlField.Kind() == reflect.String {
			str := baseControllerBaseUrlField.String()
			if str != "" {
				return str
			}
		}
	}
	baseUrlField := reflectValue.FieldByName("BaseUrl")
	if baseUrlField.IsValid() && baseUrlField.Kind() == reflect.String {
		return baseUrlField.String()
	}
	return ""
}

// newControllerFactory	为每个请求创建一个独立的控制器副本
// 控制器中的依赖字段会从注册时传入的原型复制，BaseController.Ctx 则在请求时注入
//
// param:
//   - controller	控制器原型
//
// return:
//   - 控制器工厂函数
func newControllerFactory(controller interface{}) func() reflect.Value {
	prototype := reflect.ValueOf(controller)
	if !prototype.IsValid() {
		panic("yuhuo/mvc: controller must not be nil")
	}

	controllerType := prototype.Type()
	if controllerType.Kind() == reflect.Ptr {
		if controllerType.Elem().Kind() != reflect.Struct {
			panic("yuhuo/mvc: controller must be a struct or pointer to struct")
		}
		if prototype.IsNil() {
			return func() reflect.Value { return reflect.New(controllerType.Elem()) }
		}

		prototype = prototype.Elem()
	} else if controllerType.Kind() != reflect.Struct {
		panic("yuhuo/mvc: controller must be a struct or pointer to struct")
	}

	return func() reflect.Value {
		instance := reflect.New(prototype.Type())
		instance.Elem().Set(prototype)
		return instance
	}
}

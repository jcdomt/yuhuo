// 我们提供了一种多APP形式的代码维护
// 类似于 ThinkPHP、Laravel、SpringBoot 等框架的多应用模式
// 这种能力来自于对框架的整体设计，框架本身是一个完整的应用，应用可以在框架中注册，也可以在应用中注册应用
package framework

import (
	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/mvc"
)

type webControllerInterface interface {
	Router(r mvc.ControllerRouter)
}
type runnableControllerInterface interface {
	Run() error
}
type modelInterface interface {
	TableName() string
}

type _app_url_controller_pair struct {
	url        string
	controller webControllerInterface
}

var _app_url_controller_pairs []_app_url_controller_pair
var _app_runnable_controllers []runnableControllerInterface

// 初始化
func init() {
	_app_url_controller_pairs = make([]_app_url_controller_pair, 0)
	_app_runnable_controllers = make([]runnableControllerInterface, 0)
}

// 框架层 APP 应用
type App struct {
	Url string

	// 一个 App 可以注册多个 Controller
	Controllers []interface{}

	// 一个 App 可以注册多个 Model
	ModelList []modelInterface

	// 可以定义自己的服务
	Service interface{}
}

// RegisterApp	注册一个 App 应用
//
// param:
//   - app	应用实例
func RegisterApp(app *App) {
	// 注册一个 App 应用
	// 先对该应用的 Controller 进行扫描注册
	for _, controller := range app.Controllers {
		switch c := controller.(type) {
		case webControllerInterface:
			_app_url_controller_pairs = append(_app_url_controller_pairs, _app_url_controller_pair{
				url:        app.Url,
				controller: c,
			})
		case runnableControllerInterface:
			_app_runnable_controllers = append(_app_runnable_controllers, c)
		default:
			// 跳过
		}

	}
}

// RunFramework	以 Framework 的名义启动服务器
func RunFramework(yuhuo_app *yuhuo.Application, port string) error {
	// 开始注册申请了的 Controller
	for _, pair := range _app_url_controller_pairs {
		mvc.New(yuhuo_app.Group(pair.url)).Handle(pair.controller)
	}

	err := yuhuo_app.Run(port)
	return err
}

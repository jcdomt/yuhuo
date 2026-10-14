package yuhuo

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jcdomt/yuhuo/router"
)

// Config 保存 HTTP 服务的运行配置
type Config struct {
	// 运行地址，摸扔 :8080，即监听 0.0.0.0:8080
	Addr string
	// 读写超时，默认为 10秒
	ReadTimeout time.Duration
	// 写入超时，默认为 10秒
	WriteTimeout time.Duration
	// 空闲超时，默认为 60秒
	IdleTimeout time.Duration

	// 是否启用优雅关闭，默认为 true
	RunWithGracefulShutdown bool
}

// GetDefaultConfig	返回默认的 HTTP 服务配置
//
// return:
//   - 默认的 HTTP 服务配置
func GetDefaultConfig() *Config {
	return &Config{
		Addr:                    ":8080",
		ReadTimeout:             10 * time.Second,
		WriteTimeout:            10 * time.Second,
		IdleTimeout:             60 * time.Second,
		RunWithGracefulShutdown: true,
	}
}

// Yuhuo 框架的应用层结构体，负责维护整个 HTTP 服务器应用环境
type Application struct {
	// 应用组，用于组织路由和中间件
	// 可以使用 app.Group("/prefix") 来创建子子应用组
	*ApplicationGroup
	// 配置环境
	config *Config
	// 日志记录器，Yuhuo 保证每个上下文都能获取到这个唯一的日志记录器
	logger Logger
	// 路由器，Yuhuo 内部使用路由器来管理路由和中间件
	router *router.Router

	// HTTP 服务实例，Yuhuo 内部使用这个服务来启动 HTTP 服务器
	// 最终的 HTTP 服务是由 net/http 包提供的标准服务
	server *http.Server
}

// New	创建一个新的应用实例
//
// return:
//   - 新的应用实例
func New() *Application {
	app := &Application{
		config: GetDefaultConfig(),
		logger: GetDefaultLogger(),
		router: router.GetDefaultRouter(),
	}
	// 将应用的日志记录器设置到路由器中，这样路由器在处理请求时也能使用同一个日志记录器
	app.router.SetLogger(app.logger)
	// 对内来说，主Application 其实也是一个应用组，所有的路由和中间件都可以挂载在这个主应用组上
	app.ApplicationGroup = newApplicationGroup(app, app.router.Group("/"))
	// 默认挂载恢复中间件，处理链中的 panic 会被捕获并返回 500
	app.Use(Recovery())
	return app
}

// Logger	实现 context.Application 接口
//
// return:
//   - 应用的日志记录器
func (app *Application) Logger() Logger { return app.logger }

// ServeHTTP	使 Application 可以直接作为 net/http 的 Handler 使用
//
// param:
//   - w	响应写入器
//   - r	HTTP 请求
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) { app.router.ServeHTTP(w, r) }

// Run	启动 HTTP 服务
//
// param:
//   - addr	服务监听地址
//
// return:
//   - 启动服务过程中的错误，正常关闭返回 nil
func (app *Application) Run(addr string) error {
	app.config.Addr = addr

	// 如果配置了优雅关闭，则调用 RunWithGracefulShutdown 方法启动服务
	if app.config.RunWithGracefulShutdown {
		return app.RunWithGracefulShutdown(app.config.Addr)
	}

	app.server = &http.Server{Addr: app.config.Addr, Handler: app.router, ReadTimeout: app.config.ReadTimeout, WriteTimeout: app.config.WriteTimeout, IdleTimeout: app.config.IdleTimeout}
	app.logger.Info("启动服务器：", app.config.Addr)
	err := app.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown	关闭 HTTP 服务
//
// param:
//   - ctx	关闭上下文，用于控制关闭超时
//
// return:
//   - 关闭过程中的错误
func (app *Application) Shutdown(ctx context.Context) error {
	if app.server == nil {
		return nil
	}
	app.logger.Info("服务器已关闭")
	return app.server.Shutdown(ctx)
}

// RunWithGracefulShutdown	启动服务并监听退出信号，收到信号后优雅关闭
//
// param:
//   - addr	服务监听地址
//
// return:
//   - 服务运行过程中的错误，优雅关闭时返回 nil
func (app *Application) RunWithGracefulShutdown(addr string) error {
	app.config.Addr = addr
	app.server = &http.Server{Addr: app.config.Addr, Handler: app.router, ReadTimeout: app.config.ReadTimeout, WriteTimeout: app.config.WriteTimeout, IdleTimeout: app.config.IdleTimeout}
	app.logger.Info("启动服务器：", app.config.Addr)

	errCh := make(chan error, 1)
	go func() {
		if err := app.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		app.logger.Info("收到退出信号：", sig.String())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Shutdown(ctx); err != nil {
		return err
	}
	return nil
}

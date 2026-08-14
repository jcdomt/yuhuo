package yuhuo

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jcdomt/yuhuo/router"
)

// Config 保存 HTTP 服务的运行配置。
type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// GetDefaultConfig 返回一份可直接使用的默认配置。
func GetDefaultConfig() *Config {
	return &Config{Addr: ":8080", ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
}

// Application 是应用的具体实现，负责持有路由器、日志器和 HTTP 服务。
type Application struct {
	*ApplicationGroup
	config *Config
	logger Logger
	router *router.Router
	server *http.Server
}

// New 创建一个新的应用实例。
func New() *Application {
	app := &Application{
		config: GetDefaultConfig(),
		logger: GetDefaultLogger(),
		router: router.GetDefaultRouter(),
	}
	app.router.SetLogger(app.logger)
	app.ApplicationGroup = newApplicationGroup(app, app.router.Group("/"))
	return app
}

// Logger 实现 context.Application 接口。
func (app *Application) Logger() Logger { return app.logger }

// ServeHTTP 使 Application 可以直接作为 net/http 的 Handler 使用。
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) { app.router.ServeHTTP(w, r) }

// Run 启动 HTTP 服务。
func (app *Application) Run(addr string) error {
	app.config.Addr = addr
	app.server = &http.Server{Addr: app.config.Addr, Handler: app.router, ReadTimeout: app.config.ReadTimeout, WriteTimeout: app.config.WriteTimeout, IdleTimeout: app.config.IdleTimeout}
	app.logger.Info("启动服务器：", app.config.Addr)
	err := app.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown 优关闭 HTTP 服务。
func (app *Application) Shutdown(ctx context.Context) error {
	if app.server == nil {
		return nil
	}
	return app.server.Shutdown(ctx)
}

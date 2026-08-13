package yuhuo

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jcdomt/yuhuo/router"
)

type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

func GetDefaultConfig() *Config {
	return &Config{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

type Application struct {
	*ApplicationGroup

	config *Config
	logger Logger
	router *router.Router
	server *http.Server
}

func New() *Application {
	app := &Application{
		config: GetDefaultConfig(),
		logger: GetDefaultLogger(),
		router: router.GetDefaultRouter(),
	}
	app.ApplicationGroup = newApplicationGroup(app, app.router.Group("/"))
	return app
}

func (app *Application) Run(addr string) error {
	app.config.Addr = addr
	app.server = &http.Server{
		Addr:         app.config.Addr,
		Handler:      app.router,
		ReadTimeout:  app.config.ReadTimeout,
		WriteTimeout: app.config.WriteTimeout,
		IdleTimeout:  app.config.IdleTimeout,
	}

	app.logger.Info("Starting server on ", app.config.Addr)

	err := app.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (app *Application) Shutdown(ctx context.Context) error {
	if app.server == nil {
		return nil
	}
	return app.server.Shutdown(ctx)
}

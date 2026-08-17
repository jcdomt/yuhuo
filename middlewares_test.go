package yuhuo_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/middlewares"
)

func TestCORSMiddlewareSetsHeaders(t *testing.T) {
	app := yuhuo.New()
	app.Use(middlewares.CORS(middlewares.CORSConfig{AllowOrigins: []string{"*"}}))
	app.GET("/ping", func(ctx *yuhuo.Context) { ctx.String("pong") })

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ping", nil))

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Allow-Origin = %q", got)
	}
	if response.Body.String() != "pong" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestCORSMiddlewareHandlesPreflight(t *testing.T) {
	app := yuhuo.New()
	app.Use(middlewares.CORS(middlewares.CORSConfig{
		AllowOrigins: []string{"https://example.com"},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
		MaxAge:       time.Hour,
	}))
	app.GET("/ping", func(ctx *yuhuo.Context) { ctx.String("pong") })

	request := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST" {
		t.Fatalf("Allow-Methods = %q", got)
	}
	if got := response.Header().Get("Access-Control-Max-Age"); got != "3600" {
		t.Fatalf("Max-Age = %q", got)
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	app := yuhuo.New()
	app.Use(middlewares.RequestID())
	app.GET("/id", func(ctx *yuhuo.Context) {
		ctx.String(middlewares.GetRequestID(ctx))
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/id", nil))

	id := response.Body.String()
	if id == "" {
		t.Fatal("request id 不应为空")
	}
	if got := response.Header().Get(middlewares.RequestIDHeader); got != id {
		t.Fatalf("header id = %q, body id = %q", got, id)
	}
}

func TestRequestIDMiddlewareReusesIncomingHeader(t *testing.T) {
	app := yuhuo.New()
	app.Use(middlewares.RequestID())
	app.GET("/id", func(ctx *yuhuo.Context) {
		ctx.String(middlewares.GetRequestID(ctx))
	})

	request := httptest.NewRequest(http.MethodGet, "/id", nil)
	request.Header.Set(middlewares.RequestIDHeader, "abc-123")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Body.String() != "abc-123" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestLoggerMiddlewareRecordsRequest(t *testing.T) {
	app := yuhuo.New()
	app.Use(middlewares.Logger())
	app.GET("/log", func(ctx *yuhuo.Context) { ctx.String("ok") })

	output := captureStdout(t, func() {
		response := httptest.NewRecorder()
		app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/log", nil))
	})

	if !strings.Contains(output, "GET /log") {
		t.Fatalf("日志应包含请求信息, got %q", output)
	}
}

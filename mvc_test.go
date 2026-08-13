package yuhuo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/handler"
	"github.com/jcdomt/yuhuo/mvc"
)

type testController struct{}

func (testController) Router(router mvc.ControllerRouter) {
	router.Group("/v1", func(router mvc.ControllerRouter) {
		router.GET("/health", func() interface{} {
			return handler.M{"status": "ok"}
		})
	})
}

func TestMVCApplicationHandleController(t *testing.T) {
	app := yuhuo.New()
	mvc.New(app.Group("/api")).Handle(testController{})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", body, `{"status":"ok"}`)
	}
}

type contextController struct {
	mvc.BaseController
}

func (contextController) Router(router mvc.ControllerRouter) {
	router.GET("/context", contextController{}.Value, func(ctx *yuhuo.Context) {
		ctx.Set("request-id", "abc-123")
		ctx.Next()
	})
}

func (controller contextController) Value() interface{} {
	return handler.M{"requestID": controller.Ctx.Get("request-id")}
}

func TestControllerMethodReceivesRequestContext(t *testing.T) {
	app := yuhuo.New()
	mvc.New(app.Group("/api")).Handle(contextController{})

	request := httptest.NewRequest(http.MethodGet, "/api/context", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != `{"requestID":"abc-123"}` {
		t.Fatalf("body = %q, want %q", body, `{"requestID":"abc-123"}`)
	}
}

func TestControllerFuncToHandlerFuncRejectsNil(t *testing.T) {
	app := yuhuo.New()
	router := mvc.New(app.Group("/api"))

	defer func() {
		if recover() == nil {
			t.Fatal("注册 nil 控制器函数时没有触发 panic")
		}
	}()

	router.Handle(controllerWithNilRoute{})
}

func TestControllerRouterGroupRejectsNilCallback(t *testing.T) {
	app := yuhuo.New()

	defer func() {
		if recover() == nil {
			t.Fatal("注册 nil 路由组回调时没有触发 panic")
		}
	}()

	mvc.New(app.Group("/api")).Handle(controllerWithNilGroupCallback{})
}

type controllerWithNilRoute struct{}

func (controllerWithNilRoute) Router(router mvc.ControllerRouter) {
	router.GET("/nil", nil)
}

type controllerWithNilGroupCallback struct{}

func (controllerWithNilGroupCallback) Router(router mvc.ControllerRouter) {
	router.Group("/v1", nil)
}

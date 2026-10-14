package yuhuo_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jcdomt/yuhuo"
	"github.com/jcdomt/yuhuo/framework"
)

func TestFrameworkBindJSONAndValidate(t *testing.T) {
	app := yuhuo.New()
	type payload struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age" validate:"min=1"`
	}
	app.POST("/v", func(ctx *yuhuo.Context) {
		var req payload
		if errs := framework.BindJSONAndValidate(ctx, &req); errs != nil {
			ctx.JSONs(http.StatusBadRequest, yuhuo.M{"errors": errs})
			return
		}
		ctx.JSON(yuhuo.M{"name": req.Name})
	})

	invalid := httptest.NewRecorder()
	app.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, "/v", strings.NewReader(`{}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", invalid.Code, http.StatusBadRequest)
	}
	if body := invalid.Body.String(); !strings.Contains(body, "name") {
		t.Fatalf("body 应包含校验错误, got %q", body)
	}

	valid := httptest.NewRecorder()
	app.ServeHTTP(valid, httptest.NewRequest(http.MethodPost, "/v", strings.NewReader(`{"name":"alice","age":18}`)))
	if valid.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", valid.Code, http.StatusOK)
	}
	if body := valid.Body.String(); body != `{"name":"alice"}` {
		t.Fatalf("body = %q", body)
	}
}

func TestFrameworkBindJSONAndValidateAppliesDefaults(t *testing.T) {
	app := yuhuo.New()
	type payload struct {
		Name string `json:"name" validate:"required"`
		Age  int    `json:"age" validate:"min=1" default:"18"`
	}
	app.POST("/v", func(ctx *yuhuo.Context) {
		var req payload
		if errs := framework.BindJSONAndValidate(ctx, &req); errs != nil {
			ctx.JSONs(http.StatusBadRequest, yuhuo.M{"errors": errs})
			return
		}
		ctx.JSON(yuhuo.M{"name": req.Name, "age": req.Age})
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v", strings.NewReader(`{"name":"alice"}`)))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != `{"age":18,"name":"alice"}` {
		t.Fatalf("body = %q, want 默认值注入后 {\"age\":18,\"name\":\"alice\"}", body)
	}
}

func TestFrameworkBindJSONAndValidateBadJSON(t *testing.T) {
	app := yuhuo.New()
	type payload struct {
		Name string `json:"name" validate:"required"`
	}
	app.POST("/v", func(ctx *yuhuo.Context) {
		var req payload
		if errs := framework.BindJSONAndValidate(ctx, &req); errs != nil {
			ctx.JSONs(http.StatusBadRequest, yuhuo.M{"errors": errs})
			return
		}
		ctx.JSON(yuhuo.M{"name": req.Name})
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v", strings.NewReader(`{invalid`)))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if body := response.Body.String(); !strings.Contains(body, "请求体解析失败") {
		t.Fatalf("body 应包含解析错误, got %q", body)
	}
}

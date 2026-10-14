package yuhuo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcdomt/yuhuo"
)

func TestContextBindQuery(t *testing.T) {
	app := yuhuo.New()
	type query struct {
		Name string   `form:"name"`
		Age  int      `form:"age"`
		VIP  bool     `form:"vip"`
		Rate float64  `form:"rate"`
		Tags []string `form:"tag"`
		Page int      `query:"page"`
		Skip string   `form:"-"`
	}
	app.GET("/bind", func(ctx *yuhuo.Context) {
		var q query
		if err := ctx.BindQuery(&q); err != nil {
			ctx.JSONs(http.StatusBadRequest, yuhuo.M{"error": err.Error()})
			return
		}
		ctx.JSON(yuhuo.M{
			"name": q.Name,
			"age":  q.Age,
			"vip":  q.VIP,
			"rate": q.Rate,
			"tags": q.Tags,
			"page": q.Page,
			"skip": q.Skip,
		})
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/bind?name=alice&age=30&vip=true&rate=1.5&tag=a&tag=b&page=2&skip=x", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	want := `{"age":30,"name":"alice","page":2,"rate":1.5,"skip":"","tags":["a","b"],"vip":true}`
	if body := response.Body.String(); body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

func TestContextBindQueryFieldNameFallback(t *testing.T) {
	app := yuhuo.New()
	type query struct {
		Name string
		Age  int
	}
	app.GET("/bind", func(ctx *yuhuo.Context) {
		var q query
		if err := ctx.BindQuery(&q); err != nil {
			ctx.JSONs(http.StatusBadRequest, yuhuo.M{"error": err.Error()})
			return
		}
		ctx.JSON(yuhuo.M{"name": q.Name, "age": q.Age})
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/bind?Name=alice&Age=30", nil))

	if body := response.Body.String(); body != `{"age":30,"name":"alice"}` {
		t.Fatalf("body = %q", body)
	}
}

func TestContextBindQueryConversionError(t *testing.T) {
	app := yuhuo.New()
	type query struct {
		Age int `form:"age"`
	}
	app.GET("/bind", func(ctx *yuhuo.Context) {
		var q query
		if err := ctx.BindQuery(&q); err != nil {
			ctx.JSONs(http.StatusBadRequest, yuhuo.M{"error": err.Error()})
			return
		}
		ctx.JSON(yuhuo.M{"age": q.Age})
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/bind?age=notanumber", nil))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

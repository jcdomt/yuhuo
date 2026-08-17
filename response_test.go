package yuhuo_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jcdomt/yuhuo"
)

func TestContextString(t *testing.T) {
	app := yuhuo.New()
	app.GET("/s", func(ctx *yuhuo.Context) { ctx.String("hello") })

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/s", nil))

	if ct := response.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Fatalf("content-type = %q", ct)
	}
	if response.Body.String() != "hello" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestContextHTML(t *testing.T) {
	app := yuhuo.New()
	app.GET("/h", func(ctx *yuhuo.Context) { ctx.HTML("<h1>hi</h1>") })

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/h", nil))

	if ct := response.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Fatalf("content-type = %q", ct)
	}
	if response.Body.String() != "<h1>hi</h1>" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestContextData(t *testing.T) {
	app := yuhuo.New()
	app.GET("/d", func(ctx *yuhuo.Context) { ctx.Data([]byte{0x01, 0x02, 0x03}) })

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/d", nil))

	if ct := response.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("content-type = %q", ct)
	}
	if got := response.Body.Bytes(); len(got) != 3 || got[0] != 0x01 || got[2] != 0x03 {
		t.Fatalf("body = %v", got)
	}
}

func TestContextRedirect(t *testing.T) {
	app := yuhuo.New()
	app.GET("/r", func(ctx *yuhuo.Context) { ctx.Redirect("/target") })
	app.GET("/r301", func(ctx *yuhuo.Context) { ctx.RedirectWithStatus(http.StatusMovedPermanently, "/target") })

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/r", nil))
	if response.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusFound)
	}
	if loc := response.Header().Get("Location"); loc != "/target" {
		t.Fatalf("location = %q", loc)
	}

	response2 := httptest.NewRecorder()
	app.ServeHTTP(response2, httptest.NewRequest(http.MethodGet, "/r301", nil))
	if response2.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want %d", response2.Code, http.StatusMovedPermanently)
	}
}

func TestContextFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, []byte("file content"), 0644); err != nil {
		t.Fatal(err)
	}

	app := yuhuo.New()
	app.GET("/file", func(ctx *yuhuo.Context) { ctx.File(path) })

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/file", nil))

	if response.Body.String() != "file content" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestContextCookie(t *testing.T) {
	app := yuhuo.New()
	app.GET("/set", func(ctx *yuhuo.Context) {
		ctx.SetCookie(&http.Cookie{Name: "token", Value: "xyz", Path: "/"})
		ctx.String("ok")
	})
	app.GET("/get", func(ctx *yuhuo.Context) {
		ctx.String(ctx.Cookie("token"))
	})
	app.GET("/clear", func(ctx *yuhuo.Context) {
		ctx.ClearCookie("token")
		ctx.String("ok")
	})

	setResponse := httptest.NewRecorder()
	app.ServeHTTP(setResponse, httptest.NewRequest(http.MethodGet, "/set", nil))
	cookies := setResponse.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "token" || cookies[0].Value != "xyz" {
		t.Fatalf("cookies = %+v", cookies)
	}

	request := httptest.NewRequest(http.MethodGet, "/get", nil)
	request.AddCookie(&http.Cookie{Name: "token", Value: "xyz"})
	getResponse := httptest.NewRecorder()
	app.ServeHTTP(getResponse, request)
	if getResponse.Body.String() != "xyz" {
		t.Fatalf("body = %q", getResponse.Body.String())
	}

	clearResponse := httptest.NewRecorder()
	app.ServeHTTP(clearResponse, httptest.NewRequest(http.MethodGet, "/clear", nil))
	cleared := clearResponse.Result().Cookies()
	if len(cleared) != 1 || cleared[0].MaxAge != -1 {
		t.Fatalf("cleared cookies = %+v", cleared)
	}
}

package yuhuo_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jcdomt/yuhuo"
)

func TestContextQueryParams(t *testing.T) {
	app := yuhuo.New()
	app.GET("/query", func(ctx *yuhuo.Context) {
		value, exists := ctx.GetQuery("id")
		ctx.JSON(yuhuo.M{
			"id":       value,
			"idExists": exists,
			"page":     ctx.DefaultQuery("page", "1"),
			"fallback": ctx.DefaultQuery("missing", "fallback"),
			"tags":     ctx.QueryArray("tag"),
			"queryMap": ctx.QueryMap(),
		})
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/query?id=42&tag=a&tag=b", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	want := `{"fallback":"fallback","id":"42","idExists":true,"page":"1","queryMap":{"id":["42"],"tag":["a","b"]},"tags":["a","b"]}`
	if body := response.Body.String(); body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

func TestContextPostFormURLEncoded(t *testing.T) {
	app := yuhuo.New()
	app.POST("/form", func(ctx *yuhuo.Context) {
		value, exists := ctx.GetPostForm("name")
		ctx.JSON(yuhuo.M{
			"form":    ctx.Form("name"),
			"post":    ctx.PostForm("name"),
			"exists":  exists,
			"name":    value,
			"age":     ctx.DefaultPostForm("age", "18"),
			"tags":    ctx.PostFormArray("tag"),
			"formMap": ctx.FormMap(),
		})
	})

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/form", strings.NewReader("name=alice&tag=a&tag=b"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.ServeHTTP(response, request)

	want := `{"age":"18","exists":true,"form":"alice","formMap":{"name":["alice"],"tag":["a","b"]},"name":"alice","post":"alice","tags":["a","b"]}`
	if body := response.Body.String(); body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

func TestContextPostFormMultipart(t *testing.T) {
	app := yuhuo.New()
	app.POST("/multipart", func(ctx *yuhuo.Context) {
		ctx.JSON(yuhuo.M{
			"form": ctx.Form("name"),
			"post": ctx.PostForm("name"),
		})
	})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.WriteField("name", "bob"); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/multipart", &buf)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	want := `{"form":"bob","post":"bob"}`
	if body := response.Body.String(); body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

func TestContextBindJSON(t *testing.T) {
	app := yuhuo.New()
	app.POST("/bind", func(ctx *yuhuo.Context) {
		var payload struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		if err := ctx.BindJSON(&payload); err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		ctx.JSON(yuhuo.M{"name": payload.Name, "age": payload.Age})
	})

	valid := httptest.NewRecorder()
	app.ServeHTTP(valid, httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{"name":"alice","age":30}`)))
	if body := valid.Body.String(); body != `{"age":30,"name":"alice"}` {
		t.Fatalf("body = %q", body)
	}

	invalid := httptest.NewRecorder()
	app.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(`{invalid`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", invalid.Code, http.StatusBadRequest)
	}
}

func TestContextBodyCanBeReadThenBound(t *testing.T) {
	app := yuhuo.New()
	app.POST("/raw", func(ctx *yuhuo.Context) {
		raw := ctx.BodyString()
		var payload struct {
			Name string `json:"name"`
		}
		err := ctx.BindJSON(&payload)
		ctx.JSON(yuhuo.M{"raw": raw, "name": payload.Name, "bindErr": err != nil})
	})

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/raw", strings.NewReader(`{"name":"alice"}`)))

	want := `{"bindErr":false,"name":"alice","raw":"{\"name\":\"alice\"}"}`
	if body := response.Body.String(); body != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

func TestContextFormFileAndSaveUploadedFile(t *testing.T) {
	dir := t.TempDir()
	app := yuhuo.New()
	app.POST("/upload", func(ctx *yuhuo.Context) {
		file, header, err := ctx.FormFile("file")
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		file.Close()

		dst := filepath.Join(dir, "uploaded.txt")
		if err := ctx.SaveUploadedFile(header, dst); err != nil {
			ctx.InternalServerError()
			return
		}
		ctx.JSON(yuhuo.M{"filename": header.Filename, "size": header.Size})
	})

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("hello upload")); err != nil {
		t.Fatal(err)
	}
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/upload", &buf)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if body := response.Body.String(); body != `{"filename":"hello.txt","size":12}` {
		t.Fatalf("body = %q", body)
	}

	content, err := os.ReadFile(filepath.Join(dir, "uploaded.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello upload" {
		t.Fatalf("saved content = %q", content)
	}
}

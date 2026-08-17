package yuhuo_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jcdomt/yuhuo"
)

func TestStaticFileServing(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.css"), []byte("body{}"), 0644); err != nil {
		t.Fatal(err)
	}

	app := yuhuo.New()
	app.Static("/assets", dir)

	response := httptest.NewRecorder()
	app.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/assets/app.css", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d", response.Code)
	}
	if response.Body.String() != "body{}" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

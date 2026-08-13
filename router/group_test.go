package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteGroupCombinesPrefixesAndInheritsMiddlewares(t *testing.T) {
	r := GetDefaultRouter()
	api := r.Group("/api")
	api.Use(wrapResponse("api"))

	users := api.Group("/users")
	users.Use(wrapResponse("users"))
	if err := users.Get("/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(Param(req, "id")))
	})); err != nil {
		t.Fatalf("register route: %v", err)
	}

	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/users/42", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != "api>users>42<users<api" {
		t.Fatalf("body = %q, want %q", body, "api>users>42<users<api")
	}
}

func wrapResponse(name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			_, _ = w.Write([]byte(name + ">"))
			next.ServeHTTP(w, req)
			_, _ = w.Write([]byte("<" + name))
		})
	}
}

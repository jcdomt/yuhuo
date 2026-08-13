package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterMatchesStaticRouteBeforeParameterRoute(t *testing.T) {
	r := GetDefaultRouter()
	r.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("parameter:" + Param(req, "id")))
	}))
	r.Get("/users/new", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte("static"))
	}))

	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/users/new", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if body := response.Body.String(); body != "static" {
		t.Fatalf("body = %q, want %q", body, "static")
	}
}

func TestRouterExtractsParametersAndCatchAll(t *testing.T) {
	r := GetDefaultRouter()
	r.Get("/users/:id/posts/:postID", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(Param(req, "id") + "/" + Param(req, "postID")))
	}))
	r.Get("/assets/*path", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = w.Write([]byte(Param(req, "path")))
	}))

	parameterResponse := httptest.NewRecorder()
	r.ServeHTTP(parameterResponse, httptest.NewRequest(http.MethodGet, "/users/42/posts/7", nil))
	if body := parameterResponse.Body.String(); body != "42/7" {
		t.Fatalf("parameter body = %q, want %q", body, "42/7")
	}

	catchAllResponse := httptest.NewRecorder()
	r.ServeHTTP(catchAllResponse, httptest.NewRequest(http.MethodGet, "/assets/css/site.css", nil))
	if body := catchAllResponse.Body.String(); body != "css/site.css" {
		t.Fatalf("catch-all body = %q, want %q", body, "css/site.css")
	}
}

func TestRouterReturnsMethodNotAllowedForKnownPath(t *testing.T) {
	r := GetDefaultRouter()
	r.Get("/users/:id", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/users/42", nil))

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
	if allow := response.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("Allow = %q, want %q", allow, http.MethodGet)
	}
}

func TestRouterReturnsNotFoundForUnknownPath(t *testing.T) {
	r := GetDefaultRouter()
	r.Get("/users", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	response := httptest.NewRecorder()
	r.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

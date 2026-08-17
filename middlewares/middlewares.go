package middlewares

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	requestcontext "github.com/jcdomt/yuhuo/context"
)

// RequestIDHeader 是请求标识在响应头与上下文中的键名。
const RequestIDHeader = "X-Request-Id"

// CORSConfig 是跨域中间件的配置。
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
}

// CORS 返回处理跨域请求的中间件。
func CORS(config CORSConfig) requestcontext.HandlerFunc {
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions}
	}
	allowOrigin := strings.Join(config.AllowOrigins, ", ")
	allowMethods := strings.Join(config.AllowMethods, ", ")

	return func(ctx *requestcontext.Context) {
		header := ctx.Response().Header()
		if allowOrigin != "" {
			header.Set("Access-Control-Allow-Origin", allowOrigin)
		}
		if len(config.AllowHeaders) > 0 {
			header.Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}
		if len(config.ExposeHeaders) > 0 {
			header.Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}
		if config.AllowCredentials {
			header.Set("Access-Control-Allow-Credentials", "true")
		}

		if ctx.Request().Method == http.MethodOptions {
			header.Set("Access-Control-Allow-Methods", allowMethods)
			if config.MaxAge > 0 {
				header.Set("Access-Control-Max-Age", strconv.FormatInt(int64(config.MaxAge/time.Second), 10))
			}
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}

// RequestID 为每个请求生成唯一标识，写入上下文并通过响应头返回。
func RequestID() requestcontext.HandlerFunc {
	return func(ctx *requestcontext.Context) {
		id := ctx.Request().Header.Get(RequestIDHeader)
		if id == "" {
			id = generateID()
		}
		ctx.Set(RequestIDHeader, id)
		ctx.Response().Header().Set(RequestIDHeader, id)
		ctx.Next()
	}
}

// GetRequestID 从请求上下文中取出请求标识。
func GetRequestID(ctx *requestcontext.Context) string {
	if id, ok := ctx.Get(RequestIDHeader).(string); ok {
		return id
	}
	return ""
}

// Logger 记录每个请求的方法、路径、状态码与耗时。
func Logger() requestcontext.HandlerFunc {
	return func(ctx *requestcontext.Context) {
		start := time.Now()
		ctx.Next()
		ctx.GetLogger().Info(ctx.Request().Method, " ", ctx.Request().URL.Path, " ", ctx.Status(), " ", time.Since(start))
	}
}

func generateID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(buf)
}

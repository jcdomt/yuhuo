package main

import (
	"net/http"

	"github.com/jcdomt/yuhuo"
)

func main() {
	app := yuhuo.New()
	app.Logger().SetLevel(yuhuo.LogLevelDebug)

	app.GET("/set-cookie", func(ctx *yuhuo.Context) {
		// 设置 Cookie
		ctx.SetCookie(&http.Cookie{
			Name:  "my_cookie",
			Value: "cookie_value_123",
			Path:  "/",
		})
	})

	app.GET("/get-cookie", func(ctx *yuhuo.Context) {
		// 获取 Cookie
		cookie, err := ctx.GetCookie("my_cookie")
		if err != nil {
			ctx.JSONs(400, yuhuo.M{"error": err.Error(), "code": 400})
			return
		}
		ctx.JSON(yuhuo.M{"cookie": cookie.Value})
	})

	app.GET("/clear-cookie", func(ctx *yuhuo.Context) {
		// 清除 Cookie
		ctx.ClearCookie("my_cookie")
		ctx.JSON(yuhuo.M{"message": "Cookie cleared"})
	})

	app.Run(":8080")
}

package main

import (
	"time"

	"github.com/jcdomt/yuhuo"
)

func main() {
	app := yuhuo.New()

	middleware := func(ctx *yuhuo.Context) {
		ctx.Set("middleware", "This is a middleware")

		beginTime := time.Now()
		ctx.Set("beginTime", beginTime)
		ctx.Next()
	}
	app.Use(middleware)

	app.GET("/user/:id", func(ctx *yuhuo.Context) {
		ctx.JSON(yuhuo.M{
			"message":    "Hello, World!",
			"data":       ctx.Param("id"),
			"middleware": ctx.Get("middleware"),
		})
	})

	appGroup := app.Group("/group")
	appGroup.GET("/hello", func(ctx *yuhuo.Context) {
		ctx.JSON(yuhuo.M{
			"message": "Hello, Group!",
		})
	})

	app.Use(func(ctx *yuhuo.Context) {
		endTime := time.Now()
		beginTime := ctx.Get("beginTime").(time.Time)
		duration := endTime.Sub(beginTime)
		ctx.GetLogger().Info("Request duration: ", duration)

		ctx.Next()
	})

	app.Run(":8080")
}

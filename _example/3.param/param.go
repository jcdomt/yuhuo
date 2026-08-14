package main

import "github.com/jcdomt/yuhuo"

func main() {
	app := yuhuo.New()

	app.GET("/", func(ctx *yuhuo.Context) {
		key := ctx.Query("key")
		ctx.JSON(yuhuo.M{"message": "Hello, GET!", "key": key})
	})

	app.POST("/", func(ctx *yuhuo.Context) {
		req := new(struct {
			A int `json:"a"`
			B int `json:"b"`
		})
		err := ctx.BindJSON(req)
		if err != nil {
			ctx.JSONs(400, yuhuo.M{"error": err.Error(), "code": 400})
			return
		}
		ctx.GetLogger().Info("Received POST request with data: ", req)
		ctx.JSON(yuhuo.M{"message": "Hello, POST!", "data": req})
	})

	app.GET("/:id", func(ctx *yuhuo.Context) {
		id := ctx.Param("id")
		ctx.JSON(yuhuo.M{"message": "Hello, URL!", "id": id})
	})

	app.Run(":8080")
}

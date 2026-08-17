package yuhuo

import "net/http"

// Static 将磁盘目录映射到路由前缀，自动处理目录中的静态资源。
// 例如 app.Static("/assets", "./public") 后访问 /assets/css/app.css 会读取 ./public/css/app.css。
func (group *ApplicationGroup) Static(prefix, dir string) {
	fileServer := http.StripPrefix(prefix, http.FileServer(http.Dir(dir)))
	group.GET(prefix+"/*filepath", func(ctx *Context) {
		fileServer.ServeHTTP(ctx.Response(), ctx.Request())
	})
}

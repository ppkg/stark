package auth

import (
	"github.com/go-spring/spring-core/web"
)

// 请求日志过滤器
func RequestLoggerFilter() web.Filter {
	return web.FuncFilter(func(ctx web.Context, chain web.FilterChain) {
		// start := time.Now()
		chain.Next(ctx)
		// r := ctx.Request()
		// w := ctx.ResponseWriter()
		// cost := time.Since(start)
		// log.Ctx(ctx.Context()).Infof("%s %s %s %d %d %s", r.Method, r.RequestURI, cost, w.Size(), w.Status(), r.UserAgent())
	})
}

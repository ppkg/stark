package filter

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/go-spring/spring-base/log"
	"github.com/go-spring/spring-core/web"
)

// 请求日志过滤器
func RequestLoggerFilter() web.Filter {
	return web.FuncFilter(func(ctx web.Context, chain web.FilterChain) {
		start := time.Now()
		chain.Next(ctx)
		r := ctx.Request()
		w := ctx.ResponseWriter()
		
		cost := time.Since(start)
		params := make(map[string]string)
		for k, v := range r.PostForm {
			params[k] = strings.Join(v, ",")
			
		}
		data, _ := json.Marshal(params)
		log.Ctx(ctx.Context()).Infof("%s %s %s %d %d %s %s", r.Method, r.RequestURI, cost, w.Size(), w.Status(), r.UserAgent(), data)
	})
}

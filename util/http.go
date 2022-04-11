package util

import (
	"github.com/go-spring/spring-core/web"
	"github.com/ppkg/stark/dto"
)

func ResponseJSON(ctx web.Context, data dto.HttpResponse) {
	ctx.SetContentType(web.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatus(int(data.StatusCode()))
	ctx.JSON(data)
}

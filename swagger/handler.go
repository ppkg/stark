package swagger

import (
	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/gs/cond"
	"github.com/go-spring/spring-core/web"
	SpringSwagger "github.com/go-spring/spring-swag"
	"github.com/go-spring/spring-swag/swagger"
)

func init() {
	gs.Provide(injectSwagger, "", "").On(cond.OnBean((*Document)(nil)))
	gs.Object(new(swaggerHttpServer)).Init(func(server *swaggerHttpServer) {
		gs.GetMapping("/swagger/doc.json", server.Api)
		gs.GetMapping("/swagger/*", WrapHandler)
	})
}

type swaggerHttpServer struct {
	swagger *SpringSwagger.Swagger `autowire:""`
}

func (s *swaggerHttpServer) Api(ctx web.Context) {
	ctx.SetContentType("application/json; charset=UTF-8")
	ctx.ResponseWriter().Write([]byte((s.swagger.ReadDoc())))
}

func injectSwagger(server web.Server, doc *Document) *SpringSwagger.Swagger {
	web.RegisterSwaggerHandler(func(r web.Router, doc string) {})

	rootSW := swagger.Doc(server).
		WithDescription(doc.Description).
		WithVersion(doc.Version).
		WithTitle(doc.Title).
		WithTermsOfService(doc.TermsOfService).
		WithHost(doc.Host).
		WithBasePath(doc.BasePath).
		WithSchemes(doc.Schemes...)
	for k, v := range doc.ApiKeySecurityDefinition {
		rootSW.AddApiKeySecurityDefinition(k, v)
	}
	return rootSW
}

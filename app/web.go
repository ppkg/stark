package app

import (
	"errors"

	"github.com/go-spring/spring-base/log"
	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/web/option"
	_ "github.com/go-spring/starter-echo"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/ppkg/stark"
	"google.golang.org/grpc"
)

// RunWebApplication runs http and grpc application.
func RunWebApplication(application *stark.WebApplication) {
	if application == nil || application.Application == nil {
		panic("webApplication is nil or application is nil")
	}
	// app instance once validate
	err := appInstanceOnceValidate()
	if err != nil {
		log.Errorf("禁止重复创建应用:%v", err)
		return
	}

	// 验证web应用参数
	err = validateWebApplication(application)
	if err != nil {
		log.Errorf("应用参数验证失败:%v", err)
		return
	}

	application.Type = stark.AppTypeWeb
	stark.WebInstance = application

	err = runWeb(application)
	if err != nil {
		log.Errorf("运行%s服务异常:%+v", stark.AppTypeText[application.Type], err)
	}
}

// runWeb runs http and grpc application.
func runWeb(app *stark.WebApplication) error {
	var err error

	// 注入http和grpc配置参数
	injectWebConfig(app)

	// 1. init application
	err = initApplication(app.Application)
	if err != nil {
		return err
	}

	// 2 init http and grpc vars
	err = setupWebVars(app)
	if err != nil {
		return err
	}

	return gs.Run()
}

// 验证web应用参数
func validateWebApplication(app *stark.WebApplication) error {
	if app.Port == 0 {
		return errors.New("端口号不能为空")
	}
	return nil
}

// 注入web配置参数
func injectWebConfig(app *stark.WebApplication) {
	gs.Property("spring.application.name", app.Name)
	gs.Property("web.server.port", app.Port)
}

// setupWebVars ...
func setupWebVars(app *stark.WebApplication) error {
	serverOptions := &option.GrpcServerOptions{}
	serverOptions.Options = append(serverOptions.Options, app.ServerOptions...)

	var serverUnaryInterceptors []grpc.UnaryServerInterceptor
	var serverStreamInterceptors []grpc.StreamServerInterceptor

	serverUnaryInterceptors = append(serverUnaryInterceptors, app.UnaryServerInterceptors...)
	serverStreamInterceptors = append(serverStreamInterceptors, app.StreamServerInterceptors...)

	if len(serverUnaryInterceptors) > 0 {
		serverOptions.Options = append(serverOptions.Options, grpcMiddleware.WithUnaryServerChain(serverUnaryInterceptors...))
	}
	if len(serverStreamInterceptors) > 0 {
		serverOptions.Options = append(serverOptions.Options, grpcMiddleware.WithStreamServerChain(serverStreamInterceptors...))
	}

	if len(serverOptions.Options) > 0 {
		gs.Object(serverOptions)
	}

	return nil
}

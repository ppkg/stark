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

// RunGrpcApplication runs grpc application.
func RunGrpcApplication(application *stark.GrpcApplication) {
	if application == nil || application.Application == nil {
		panic("grpcApplication is nil or application is nil")
	}
	// app instance once validate
	err := appInstanceOnceValidate()
	if err != nil {
		log.Errorf("禁止重复创建应用:%v", err)
		return
	}

	// 验证grpc应用参数
	err = validateGrpcApplication(application)
	if err != nil {
		log.Errorf("应用参数验证失败:%v", err)
		return
	}

	application.Type = stark.AppTypeGrpc
	stark.GrpcInstance = application

	err = runGrpc(application)
	if err != nil {
		log.Errorf("运行Grpc服务异常:%+v", err)
	}
}

// runGrpc runs grpc application.
func runGrpc(grpcApp *stark.GrpcApplication) error {
	var err error

	// 注入grpc配置参数
	injectGrpcConfig(grpcApp)

	// 1. init application
	err = initApplication(grpcApp.Application)
	if err != nil {
		return err
	}

	// 2 init grpc vars
	err = setupGrpcVars(grpcApp)
	if err != nil {
		return err
	}

	return gs.Run()
}

// 验证grpc应用参数
func validateGrpcApplication(grpcApp *stark.GrpcApplication) error {
	if grpcApp.Port == 0 {
		return errors.New("端口号不能为空")
	}
	return nil
}

// 注入grpc配置参数
func injectGrpcConfig(grpcApp *stark.GrpcApplication) {
	gs.Property("spring.application.name", grpcApp.Name)
	gs.Property("web.server.port", grpcApp.Port)
}

// setupGrpcVars ...
func setupGrpcVars(grpcApp *stark.GrpcApplication) error {
	serverOptions := &option.GrpcServerOptions{}
	serverOptions.Options = append(serverOptions.Options, grpcApp.ServerOptions...)

	var serverUnaryInterceptors []grpc.UnaryServerInterceptor
	var serverStreamInterceptors []grpc.StreamServerInterceptor

	serverUnaryInterceptors = append(serverUnaryInterceptors, grpcApp.UnaryServerInterceptors...)
	serverStreamInterceptors = append(serverStreamInterceptors, grpcApp.StreamServerInterceptors...)

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

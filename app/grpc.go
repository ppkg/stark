package app

import (
	"github.com/go-spring/spring-base/log"
	"github.com/ppkg/stark"
)

// RunGrpcApplication runs grpc application.
func RunGrpcApplication(application *stark.GrpcApplication) {
	if application == nil || application.Application == nil {
		panic("grpcApplication is nil or application is nil")
	}
	// app instance once validate
	{
		err := appInstanceOnceValidate()
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	application.Type = stark.AppTypeGrpc
	stark.GrpcInstance = application

	err := runGrpc(application)
	if err != nil {
		log.Errorf("grpcApp runGrpc err: %v\n", err)
	}
}

// runGrpc runs grpc application.
func runGrpc(grpcApp *stark.GrpcApplication) error {
	var err error

	// 1. init application
	err = initApplication(grpcApp.Application)
	if err != nil {
		return err
	}

	// 2 init grpc vars
	err = setupGRPCVars(grpcApp)
	if err != nil {
		return err
	}

	return nil
}

// setupGrpcVars ...
func setupGRPCVars(grpcApp *stark.GrpcApplication) error {
	// var err error

	return nil
}

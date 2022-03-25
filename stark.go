package stark

import (
	"crypto/tls"

	"google.golang.org/grpc"
)

const (
	AppTypeGrpc  = 1
	AppTypeCron  = 2
	AppTypeQueue = 3
	AppTypeHttp  = 4
)

var (
	AppTypeText = map[int32]string{
		AppTypeGrpc:  "gRPC",
		AppTypeCron:  "Cron",
		AppTypeQueue: "Queue",
		AppTypeHttp:  "Http",
	}
)

// Application ...
type Application struct {
	Name        string
	Type        int32
	Environment string
	LoadConfig  func() error
	SetupVars   func() error
	StopFunc    func() error
}

// GRPCApplication ...
type GrpcApplication struct {
	*Application
	Port                     int64
	TlsConfig                *tls.Config
	UnaryServerInterceptors  []grpc.UnaryServerInterceptor
	StreamServerInterceptors []grpc.StreamServerInterceptor
	ServerOptions            []grpc.ServerOption
}

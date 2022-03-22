package option

import "google.golang.org/grpc"

type GrpcOption struct {
	// 服务端可选参数
	ServerOptions []grpc.ServerOption
	// 客户端可选参数
	ClientOptions []grpc.DialOption
}

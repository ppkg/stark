package gateway

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type GatewayServer struct {
	// grpc网关注册函数
	Register func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
}

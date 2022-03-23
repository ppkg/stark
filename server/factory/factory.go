package factory

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/go-spring/spring-base/log"
	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/grpc"
	"github.com/go-spring/spring-core/gs"
	gwRuntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ppkg/starter-grpc/server/gateway"
	"github.com/ppkg/starter-grpc/server/option"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	g "google.golang.org/grpc"
)

// Starter gRPC 服务器启动器
type Starter struct {
	config      grpc.ServerConfig
	server      *g.Server
	mux         *http.ServeMux
	gwMux       *gwRuntime.ServeMux
	servers     *gs.GrpcServers          `autowire:""`
	gwRegisters []*gateway.GatewayServer `autowire:""`
	option      *option.GrpcOption       `autowire:"?"`
}

// NewStarter Starter 的构造函数
func NewStarter(config grpc.ServerConfig) *Starter {
	return &Starter{
		config: config,
	}
}

func (starter *Starter) OnAppStart(ctx gs.Context) {
	var ops []g.ServerOption
	if starter.option != nil {
		ops = starter.option.ServerOptions
	}
	starter.server = g.NewServer(ops...)

	server := reflect.ValueOf(starter.server)
	srvMap := make(map[string]reflect.Value)

	starter.servers.ForEach(func(serviceName string, rpcServer *grpc.Server) {
		service := reflect.ValueOf(rpcServer.Service)
		srvMap[serviceName] = service
		fn := reflect.ValueOf(rpcServer.Register)
		fn.Call([]reflect.Value{server, service})
	})

	for service, info := range starter.server.GetServiceInfo() {
		srv := srvMap[service]
		for _, method := range info.Methods {
			m, _ := srv.Type().MethodByName(method.Name)
			fnPtr := m.Func.Pointer()
			fnInfo := runtime.FuncForPC(fnPtr)
			file, line := fnInfo.FileLine(fnPtr)
			log.Infof("/%s/%s %s:%d ", service, method.Name, file, line)
		}
	}

	// 初始化http多路复用器
	starter.mux = http.NewServeMux()
	starter.gwMux = gwRuntime.NewServeMux()
	// 注册grpc-gateway
	starter.registerGrpcGateway()

	ctx.Go(func(_ context.Context) {
		if err := starter.doServe(); err != nil {
			log.Error(nil, err)
		}
	})
}

func (starter *Starter) OnAppStop(ctx context.Context) {
	starter.server.GracefulStop()
}

// 注册gprc gateway
func (starter *Starter) registerGrpcGateway() {
	for _, s := range starter.gwRegisters {
		dialOpts := []g.DialOption{g.WithInsecure()}
		err := s.Register(context.Background(), starter.gwMux, fmt.Sprintf("localhost:%d", starter.config.Port), dialOpts)
		util.Panic(err).When(err != nil)
	}
}

// 运行服务
func (starter *Starter) doServe() error {
	starter.mux.Handle("/", starter.gwMux)
	http.ListenAndServe(fmt.Sprintf(":%d", starter.config.Port), starter.grpcHandlerFunc())
	return nil
}

func (starter *Starter) grpcHandlerFunc() http.Handler {
	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			starter.server.ServeHTTP(w, r)
		} else {
			starter.mux.ServeHTTP(w, r)
		}
	}), &http2.Server{})
}

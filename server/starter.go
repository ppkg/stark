package StarterGrpcServer

import (
	"github.com/go-spring/spring-core/gs"
	"github.com/ppkg/starter-grpc/server/factory"
)

func init() {
	gs.Provide(factory.NewStarter, "${grpc.server}").Name("starter.grpc").Export((*gs.AppEvent)(nil))
}

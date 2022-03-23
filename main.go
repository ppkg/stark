/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-spring/spring-core/grpc"
	"github.com/go-spring/spring-core/gs"

	hello_world "github.com/ppkg/starter-grpc/proto/helloworld"
	_ "github.com/ppkg/starter-grpc/server"
	"github.com/ppkg/starter-grpc/server/gateway"
)

func init() {
	gs.Object(new(GreeterServer)).Init(func(s *GreeterServer) {
		gs.GrpcServer("helloworld.Greeter", &grpc.Server{
			Register: hello_world.RegisterGreeterServer,
			Service:  s,
		})
	})
	gs.Object(&gateway.GatewayServer{
		Register: hello_world.RegisterGreeterHandlerFromEndpoint,
	}).Name("gateway.greeter")
}

type GreeterServer struct {
	AppName string `value:"${spring.application.name}"`
}

func (s *GreeterServer) SayHello(ctx context.Context, in *hello_world.HelloRequest) (*hello_world.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &hello_world.HelloReply{Message: "Hello " + in.GetName() + " from " + s.AppName}, nil
}

func main() {
	gs.Property("spring.application.name", "GreeterServer")
	gs.Property("grpc.server.port", 50051)
	fmt.Println("application exit: ", gs.Web(false).Run())
}

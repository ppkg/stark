package stark

import (
	"crypto/tls"
	"fmt"

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

// 数据库类型
const (
	DbTypeMyql  = 1
	DbTypeRedis = 2
)

var (
	DbTypeText = map[int32]string{
		DbTypeMyql:  "MySQL",
		DbTypeRedis: "Redis",
	}
)

// 数据库连接信息
type DbConnInfo struct {
	Name string
	Url  string
	Type int32
	// 其他配置信息
	Extras map[string]interface{}
}

// Application ...
type Application struct {
	Name        string
	Type        int32
	Environment string
	IsDebug     bool
	LoadConfig  func() error
	SetupVars   func() error
	StopFunc    func() error
	dbConnInfos map[string]DbConnInfo
}

// 放入一个数据库连接信息
func (s *Application) PutDbConn(info DbConnInfo) error {
	if s.dbConnInfos == nil {
		s.dbConnInfos = make(map[string]DbConnInfo)
	}
	if _, ok := s.dbConnInfos[info.Name]; ok {
		return fmt.Errorf("Name值(%s)必须是唯一", info.Name)
	}
	s.dbConnInfos[info.Name] = info
	return nil
}

// 获取所有数据库连接信息
func (s *Application) GetDbConns() []DbConnInfo {
	list := make([]DbConnInfo, 0, len(s.dbConnInfos))
	for _, v := range s.dbConnInfos {
		list = append(list, v)
	}
	return list
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

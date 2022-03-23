package StarterGrpc

type ApplicationContext interface {
	// 初始化
	Init() error
	// 安装组件
	Setup() error
	// 执行业务
	Run() error
}

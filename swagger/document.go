package swagger

type Document struct {
	// 项目描述
	Description string
	// API版本号
	Version string
	// 服务名称
	Title string
	// 服务条款地址
	TermsOfService string
	// 可用服务器地址
	Host string
	// API 路径的前缀
	BasePath string
	// 服务协议
	Schemes []string
	// 应用ID
	Id string
	// ApiKey 方式认证,map中key是认证参数名，map中value则是参数来源，这里demo是来源请求头header,demo:
	// map[string]string{
	// 	"token":"header"
	// }
	ApiKeySecurityDefinition map[string]string
}

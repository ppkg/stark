package dto

type CommonHttpResponse struct {
	// 状态码，200：请求成功，400：请求失败
	Code int32 `json:"-"`
	// 响应提示信息
	Message string `json:"message"`
	// 业务数据
	Data interface{} `json:"data"`
}

func (s CommonHttpResponse) StatusCode() int32 {
	return s.Code
}

type CommonPageHttpResponse struct {
	// 状态码，200：请求成功，400：请求失败
	Code int32 `json:"-"`
	// 响应提示信息
	Message string `json:"message"`
	// 业务数据
	Data interface{} `json:"data"`
	// 总记录大小
	Total int32 `json:"total"`
}

func (s CommonPageHttpResponse) StatusCode() int32 {
	return s.Code
}

type HttpResponse interface {
	StatusCode() int32
}

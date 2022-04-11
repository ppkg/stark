package models

type LoginUserInfo struct {
	//
	Iss string `json:"iss"`
	//
	Aud string `json:"aud"`
	//
	Jti string `json:"jti"`
	//
	Iat string `json:"iat"`
	//
	Nbf string `json:"nbf"`
	//
	Exp string `json:"exp"`

	//用户编号
	UserNo string `json:"userNo"`
	//手机号
	Mobile string `json:"mobile"`
	//真实姓名
	RealName string `json:"realName"`
	//邮箱
	Email string `json:"email"`
	//用户名
	UserName string `json:"username"`

	//boss后台使用的手机号(Mobile去除+86字符)
	BossMobile string
	//记录当前账号是否选择了单个门店
	FinancialCode string `json:"financialCode"`
	//记录是否是总账号标识
	IsGeneralAccount bool
	//记录是否是总账号标识
	IsSkipGeneral bool
}
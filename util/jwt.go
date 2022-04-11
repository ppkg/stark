package util

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/ppkg/stark/models"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo/v4"
	"github.com/limitedlee/microservice/common/config"
	micjwt "github.com/limitedlee/microservice/common/jwt"
	"github.com/spf13/cast"
)

//创建用户jwt
func CreateJwtToken(mobile string, userno string, userName string) (tokenStr string, err error) {
	// Create token
	claims := make(jwt.MapClaims)
	claims["userno"] = userno
	claims["name"] = userName
	claims["mobile"] = mobile
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenStr, err = token.SignedString(micjwt.PrivateKey) //需要添加私钥进行加密
	return
}

// 参数tokenStr指的是 从客户端传来的待验证Token
// 验证Token过程中，如果Token生成过程中，指定了iat与exp参数值，将会自动根据时间戳进行时间验证
// 返回payload例子，map[exp:1.562839369e+09 iat:1.562234569e+09 iss:rp-pet.com mobile:18576 nameid:2 role:member]
func ParseAndGetPayload(tokenStr string) (jwt.MapClaims, error) {
	// 基于公钥验证Token合法性
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// 基于JWT的第一部分中的alg字段值进行一次验证
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("验证登录信息的加密类型错误")
		}
		return micjwt.PublicKey, nil
	},
	)
	if err != nil {
		log.Println("验证Token失败：", err, "tokenStr: ", tokenStr)
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		//判断是否过期
		nowTime := time.Now().Unix()
		jwtTime := int64(claims["exp"].(float64))
		if nowTime-jwtTime > 0 {
			return nil, errors.New("登录信息已过期")
		}
		return claims, nil
	}
	return claims, errors.New("登录信息无效或者无对应值")
}

//系统内部使用解析token
func GetPayloadDirectly(cx echo.Context) (jwt.MapClaims, error) {
	token := ""
	jwtToken := cx.Request().Header.Get("Authorization")
	if len(jwtToken) <= 0 {
		return nil, errors.New("无登录信息")
	}

	index := strings.Index(jwtToken, " ")
	count := strings.Count(jwtToken, "")
	token = jwtToken[index+1 : count-1]
	if len(token) <= 0 {
		return nil, errors.New("无登录信息")
	}

	claims, err := ParseAndGetPayload(token)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

//系统内部使用解析token
func GetPayloadDirectlyToInterface(cx echo.Context) (models.LoginUserInfo, error) {

	var userInfo models.LoginUserInfo
	token := ""
	jwtToken := cx.Request().Header.Get("Authorization")

	if len(jwtToken) <= 0 {
		return userInfo, errors.New("无登录信息")
	}

	index := strings.Index(jwtToken, " ")
	count := strings.Count(jwtToken, "")
	token = jwtToken[index+1 : count-1]

	if len(token) <= 0 {
		return userInfo, errors.New("无登录信息")
	}

	claims, err := ParseAndGetPayload(token)
	if err != nil {
		return userInfo, err
	}

	userInfo.IsGeneralAccount = false
	userInfo.UserNo = cast.ToString(claims["userno"])
	userInfo.UserName = cast.ToString(claims["name"])
	userInfo.Mobile = cast.ToString(claims["mobile"])
	//增加用户账号信息判断
	userInfo.FinancialCode = cx.Request().Header.Get("financialCode") //财务编码
	//保存信息到渠道商品【具有总帐号权限】
	adminAccount := config.GetString("adminAccountList")
	adminAccountArry := strings.Split(adminAccount, "|")
	adminAccountMap := GenerateArryToMap(adminAccountArry)
	if _, ok := adminAccountMap[userInfo.UserNo]; ok {
		userInfo.IsGeneralAccount = true
	}
	return userInfo, nil
}

// 获取当前登录用户名
func GetCurrentUser(cx echo.Context) models.LoginUserInfo {
	user, err := GetPayloadDirectlyToInterface(cx)
	if err == nil {
		return user
	}
	return models.LoginUserInfo{}
}

func ValidateToken(token string, publicKey *rsa.PublicKey) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			//glog.Printf("Unexpected signing method: %v", t.Header["alg"])
			return nil, fmt.Errorf("invalid token")
		}
		return publicKey, nil
	})
	if err == nil && jwtToken.Valid {
		return jwtToken, nil
	}
	return nil, err
}

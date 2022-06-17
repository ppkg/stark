package filter

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-spring/spring-core/web"
	"github.com/limitedlee/microservice/common/config"
	jw "github.com/limitedlee/microservice/common/jwt"
	"github.com/maybgit/glog"
	utils "github.com/ppkg/stark/util"
	"github.com/ucarion/urlpath"
)

// 授权认证过滤器
type LoginFilter struct {
	// 过滤掉不需要登录的请求，不在这里面的都需要登录
	UrlPatterns []string
	// 获取需要认证app信息
	// 认证失败回调
	AuthFail      func(ctx web.Context, status int32, err error)
	pathMatcher   []urlpath.Path
	RedisClient   *redis.Client
	isInitMatcher bool
}

type LoginInfo struct {
	// 应用ID
	LoginUrl int32 `json:"Login_url"`
}

type UserInfo struct {
	ServiceResponse struct {
		AuthenticationSuccess struct {
			User       string `json:"user"`
			Attributes struct {
				ShrCode                                []string  `json:"shrCode"`
				IsFromNewLogin                         []bool    `json:"isFromNewLogin"`
				AuthenticationDate                     []float64 `json:"authenticationDate"`
				UserNo                                 []string  `json:"userNo"`
				SuccessfulAuthenticationHandlers       []string  `json:"successfulAuthenticationHandlers"`
				Mobile                                 []string  `json:"mobile"`
				Type                                   []string  `json:"type"`
				CredentialType                         string    `json:"credentialType"`
				RealName                               []string  `json:"realName"`
				AuthenticationMethod                   string    `json:"authenticationMethod"`
				LongTermAuthenticationRequestTokenUsed []bool    `json:"longTermAuthenticationRequestTokenUsed"`
				ID                                     []int     `json:"id"`
				Email                                  []string  `json:"email"`
				Username                               []string  `json:"username"`
			} `json:"attributes"`
		} `json:"authenticationSuccess"`
	} `json:"serviceResponse"`
}

type UserInfoFailure struct {
	ServiceResponse struct {
		AuthenticationFailure struct {
			Code        string `json:"code"`
			Description string `json:"description"`
		} `json:"authenticationFailure"`
	} `json:"serviceResponse"`
}

type UserInfoAuth struct {
	UserName string      `json:"user_name"`
	Email    string      `json:"email"`
	ID       int         `json:"id"`
	Mobile   string      `json:"mobile"`
	RealName string      `json:"real_name"`
	UserNo   string      `json:"user_no"`
	UserAuth interface{} `json:"user_auth"`
}

type AuthData struct {
	//权限集合
	Operation interface{} `json:"operation"`
	//权限范围
	Resource interface{} `json:"resource"`
}

type UserAuthResponse struct {
	//编码
	Code int32 `json:"code"`
	//请求状态码
	Status int32 `json:"status"`
	//提示
	Message string `json:"message"`
	//数据
	AuthData `json:"data"`
}

//登入确认响应
type LoginResponse struct {
	//响应码
	Code int32 `json:"code"`
	//消息
	Message string `json:"message"`
	//登录口令
	Token string `json:"token"`
	//权限集
	Auths string `json:"auths"`
}

//调用北京获取用户权限接口，将结果权限集缓存redis
func GetUserAuthCollection(systemcode, memberid, structOuterCode string) (string, error) {
	url := fmt.Sprintf("%s%s", config.GetString("oms-bj-user-auth-url"), "/api/priv/list")

	url = url + "?systemCode=" + systemcode + "&userCode=" + memberid + "&companyCode=" + structOuterCode

	//2：调用北京获取权限接口
	data, err := utils.HttpGetUrlOMS(url)

	if err != nil {
		return "MapToJson转换出错", err
	}
	//if code != 200 {
	//	return "", errors.New("获取权限集失败")
	//}

	return data, nil
}

//北京acp接口公用参数
func createCommonBjAcpParam() map[string]interface{} {
	dataArr := make(map[string]interface{})

	apiid := config.GetString("BJAuth.AppId")
	apiSecret := config.GetString("BJAuth.ApiSecret")
	apiStr := utils.GenSonyflake() //自己生成，唯一的十六位随机字符串
	timestamp := strconv.Itoa(int(time.Now().Unix()))
	sign := fmt.Sprintf("apiSecret=%s&apiStr=%s&apiId=%s&timestamp=%s&apiSecret=%s", apiSecret, apiStr, apiid, timestamp, apiSecret)
	h := md5.New()
	h.Write([]byte(sign))
	md5sign := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))

	dataArr["apiId"] = apiid
	dataArr["apiStr"] = apiStr
	dataArr["timestamp"] = timestamp
	dataArr["sign"] = md5sign

	return dataArr
}

func (s *LoginFilter) Invoke(ctx web.Context, chain web.FilterChain) {

	// 如果符合不需要登录的直接跳过
	urlReturn := s.isMatch(ctx)
	omsToken := ""
	//请求的路径
	w := ctx.ResponseWriter()
	r := ctx.Request()
	switch urlReturn {
	case 1:
		chain.Next(ctx)
		return
	case 2:
		cookie, _ := ctx.Request().Cookie("oms_token")
		if cookie != nil {
			omsToken = cookie.Value
		}
		if omsToken == "" {
			//如果没有token 直接返回失败，调用反返回401
			s.AuthFail(ctx, 401, errors.New("用户没有登录！"))
			return
		} else {
			//如果 omsToken 不为空，说明已经登录过的，解析token看是否过期,没有过期的话，重新续期

			token := omsToken

			_, err := utils.ValidateToken(token, jw.PublicKey)
			if err != nil {
				s.AuthFail(ctx, 401, errors.New(fmt.Sprintf("valid token required.%v", err)))
				return
			}
			//判断token 是否失效
			claims, err := utils.ParseAndGetPayload(token)
			//获取redis进行匹配
			RedisToken, err := s.RedisClient.Get(ctx.Context(), fmt.Sprintf("oms:token:%s", claims["userno"])).Result()
			if err != nil { //redis 获取失败
				glog.Error("AuthFilter read redis err", err)
				s.AuthFail(ctx, 401, errors.New("redis获取token失败,token失效"))
				return
			}
			if RedisToken != token { //redis 不一致
				s.AuthFail(ctx, 401, errors.New("token失效"))
				return
			}

			cookie_token := new(http.Cookie)
			cookie_token.Name = "oms_token"
			cookie_token.Value = token
			cookie_token.Path = "/"
			//cookie有效期为3600秒
			cookie_token.MaxAge = 2 * 86400
			ctx.SetCookie(cookie_token)

			cookie_user := new(http.Cookie)
			cookie_user.Name = "user_no"
			cookie_user.Value = fmt.Sprintf("%s", claims["userno"])
			cookie_user.Path = "/"
			//cookie有效期为3600秒
			cookie_user.MaxAge = 2 * 86400
			ctx.SetCookie(cookie_user)

			//缓存jwt
			err = s.RedisClient.Set(ctx.Context(), fmt.Sprintf("oms:token:%s", claims["userno"]), token, time.Hour*2).Err()
			if err != nil {
				s.AuthFail(ctx, 500, errors.New("token缓存redis失败"))
				return
			}

			chain.Next(ctx)

		}
		break
	case 3:
		//如果有ticket说明是跳转cas登录后的回调请求,就可以直接获取权限，否者跳转登录界面
		ticket := s.getRequestParam(ctx, "ticket")
		Url := s.getRequestParam(ctx, "url")
		casUrl := config.GetString("cas.login.url")
		omsurl := config.GetString("oms.login.url")
		if ticket == "" {
			returnUrl := casUrl + "/cas/login?service=" + url.QueryEscape(omsurl+"?url="+url.QueryEscape(Url))
			http.Redirect(w, r, returnUrl, http.StatusFound) //跳转到登录界面
			return
		} else {

			//通过CAS登录接口验证，获取用户的手机，用户编码， 邮箱等信息
			getUserInfoUrl := casUrl + "/cas/p3/serviceValidate?format=json&ticket=" + ticket + "&service=" + url.QueryEscape(omsurl+"?url="+Url)
			data := utils.HttpGetUrl(getUserInfoUrl)
			if len(data) > 0 {
				if strings.Contains(data, "authenticationFailure") {
					info := UserInfoFailure{}
					json.Unmarshal([]byte(data), &info)
					s.AuthFail(ctx, 401, errors.New(info.ServiceResponse.AuthenticationFailure.Description))
					return
				}

				info := UserInfo{}
				json.Unmarshal([]byte(data), &info)

				user := info.ServiceResponse.AuthenticationSuccess.Attributes
				userInfoAuth := UserInfoAuth{}
				userInfoAuth.UserNo = user.UserNo[0]
				userInfoAuth.ID = user.ID[0]
				userInfoAuth.Email = user.Email[0]
				userInfoAuth.Mobile = user.Mobile[0]
				userInfoAuth.UserName = user.Username[0]
				userInfoAuth.RealName = user.RealName[0]

				//4:将权限数据存储到redis
				systemCode := config.GetString("OmsSystemCode")
				authStr, autherr := GetUserAuthCollection(systemCode, userInfoAuth.UserNo, "RPX0001")
				var baseRes = UserAuthResponse{}
				err := json.Unmarshal([]byte(authStr), &baseRes)
				if autherr != nil || err != nil {
					s.AuthFail(ctx, 500, errors.New("获取用户权限失败！"))
					return
				}
				//
				//beginIndex := strings.Index(authStr, `"operation": {`)
				//endIndex := strings.Index(authStr, `},`)
				//nowStr := authStr[beginIndex+14 : endIndex]
				//nowStr = strings.ReplaceAll(nowStr, " ", "")
				//nowStr = strings.ReplaceAll(nowStr, "\n", "")
				userInfoAuth.UserAuth = baseRes.AuthData.Operation
				jsonUserAuth, err := json.Marshal(userInfoAuth)

				//2:重新组装jwt用户身份验证
				jwtString, err := utils.CreateJwtToken(userInfoAuth.Mobile, userInfoAuth.UserNo, userInfoAuth.UserName)
				if err != nil {
					s.AuthFail(ctx, 500, errors.New("生成JWT失败！"))
					return
				}

				//缓存jwt
				err = s.RedisClient.Set(ctx.Context(), fmt.Sprintf("oms:token:%s", userInfoAuth.UserNo), jwtString, time.Hour*2).Err()
				if err != nil {
					s.AuthFail(ctx, 500, errors.New("token缓存redis失败"))
					return
				}

				cookie := new(http.Cookie)
				cookie.Name = "oms_token"
				cookie.Value = jwtString
				cookie.Path = "/"
				//cookie有效期为3600秒
				cookie.MaxAge = 2 * 86400
				ctx.SetCookie(cookie)

				cookie_user := new(http.Cookie)
				cookie_user.Name = "user_no"
				cookie_user.Value = userInfoAuth.UserNo
				cookie_user.Path = "/"
				//cookie有效期为3600秒
				cookie_user.MaxAge = 2 * 86400
				ctx.SetCookie(cookie_user)

				//缓存用户权限信息
				err = s.RedisClient.Set(ctx.Context(), fmt.Sprintf("oms:auth:%s", userInfoAuth.UserNo), jsonUserAuth, 0).Err()
				if err != nil {
					s.AuthFail(ctx, 500, errors.New("缓存用户权限失败"))
					return
				}

				http.Redirect(w, r, Url, http.StatusFound) //跳转到百度
				return
			}
		}
		break
	}
	chain.Next(ctx)
}

// 获取请求参数
func (s *LoginFilter) getRequestParam(ctx web.Context, key string) string {
	val := ctx.QueryParams().Get(key)
	if val != "" {
		return val
	}
	val = ctx.Request().PostFormValue(key)
	if val != "" {
		return val
	}
	val = ctx.Request().Header.Get(key)
	if val != "" {
		return val
	}
	return val
}

//验证是否是不需要验证登录的路由，如果是不需要登录的直接跳过
func (s *LoginFilter) isMatch(ctx web.Context) int {
	if !s.isInitMatcher {
		s.pathMatcher = make([]urlpath.Path, 0, len(s.UrlPatterns))
		for _, v := range s.UrlPatterns {
			s.pathMatcher = append(s.pathMatcher, urlpath.New(v))
		}
		s.isInitMatcher = true
	}

	uri := ctx.Request().URL.Path
	//if strings.Contains(uri, "login_ticket") {
	//	return 3
	//}
	if strings.Contains(uri, "logout") {
		return 1
	}
	for _, v := range s.pathMatcher {
		_, ok := v.Match(uri)
		if ok {
			return 1
		}
	}
	return 2
}

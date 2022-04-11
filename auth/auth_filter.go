package auth

import (
	"bytes"
	"errors"
	"io/ioutil"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ppkg/stark/util"
	"github.com/ucarion/urlpath"

	"github.com/go-spring/spring-base/cast"
	"github.com/go-spring/spring-base/log"
	"github.com/go-spring/spring-core/web"
)

// 授权应用Key
const GrantAppKey = "grantApp"

// 授权认证过滤器
type AuthFilter struct {
	// 过滤器匹配路由
	UrlPatterns []string
	// 认证成功回调
	AuthSuccess func(ctx web.Context, app AppInfo)
	// 认证失败回调
	AuthFail func(ctx web.Context, err error)
	// 获取需要认证app信息
	GetApps          func() ([]AppInfo, error)
	isInitUrlPattern bool
	pathMatcher      []urlpath.Path
}

func (s *AuthFilter) isMatch(ctx web.Context) bool {
	if !s.isInitUrlPattern {
		s.pathMatcher = make([]urlpath.Path, 0, len(s.UrlPatterns))
		for _, v := range s.UrlPatterns {
			s.pathMatcher = append(s.pathMatcher, urlpath.New(v))
		}
	}

	uri := ctx.Request().RequestURI
	for _, v := range s.pathMatcher {
		_, ok := v.Match(uri)
		if ok {
			return true
		}
	}
	return false
}

func (s *AuthFilter) Invoke(ctx web.Context, chain web.FilterChain) {
	// 如果不符合过滤规则直接跳过
	if !s.isMatch(ctx) {
		chain.Next(ctx)
		return
	}

	appId := s.getRequestParam(ctx, "appId")
	sign := s.getRequestParam(ctx, "sign")
	timestamp := s.getRequestParam(ctx, "timestamp")

	// 如果时间戳不在半个小时范围内则直接丢失
	if int(math.Abs(time.Since(time.Unix(cast.ToInt64(timestamp), 0)).Minutes())) > 30 {
		if s.AuthFail != nil {
			s.AuthFail(ctx, errors.New("时间戳已失效，有效时间为30分钟"))
		}
		return
	}

	apps, err := s.GetApps()
	if err != nil {
		log.Errorf("authFilter 获取应用信息异常:%+v appId=%s sign=%s", err, appId, sign)
		if s.AuthFail != nil {
			s.AuthFail(ctx, err)
		}
		return
	}

	myAppId := cast.ToInt32(appId)
	var myApp AppInfo
	for _, v := range apps {
		if v.Id == myAppId {
			myApp = v
			break
		}
	}

	if myApp.Id == 0 {
		log.Warnf("authFilter 应用未授权 appId=%s sign=%s", appId, sign)
		if s.AuthFail != nil {
			s.AuthFail(ctx, errors.New("应用未授权"))
		}
		return
	}

	newSign, err := s.generateSign(ctx, myApp)
	if err != nil {
		log.Errorf("authFilter 生成签名失败:%+v appId=%s sign=%s", err, appId, sign)
		if s.AuthFail != nil {
			s.AuthFail(ctx, err)
		}
		return
	}

	if sign != newSign {
		log.Errorf("authFilter 签名验证不通过 appId=%s sign=%s", appId, sign)
		if s.AuthFail != nil {
			s.AuthFail(ctx, errors.New("签名验证不通过"))
		}
		return
	}

	err = ctx.Set("grantApp", myApp)
	if err != nil {
		log.Errorf("authFilter 设置上下文数据失败:%+v appId=%s sign=%s", err, appId, sign)
		if s.AuthFail != nil {
			s.AuthFail(ctx, err)
		}
		return
	}

	if s.AuthSuccess != nil {
		s.AuthSuccess(ctx, myApp)
	}

	chain.Next(ctx)
}

// 生成签名
func (s *AuthFilter) generateSign(ctx web.Context, app AppInfo) (string, error) {
	params := make(map[string]string)
	s.mergeRequestParams(params, ctx.QueryParams())
	formParams, err := ctx.FormParams()
	if err != nil {
		log.Errorf("authFilter 获取表单参数异常:%+v", err)
		return "", err
	}
	s.mergeRequestParams(params, formParams)

	if val := s.getRequestParam(ctx, "timestamp"); val != "" {
		params["timestamp"] = val
	}

	var sign string
	contentType := ctx.Request().Header.Get("content-type")
	if strings.HasPrefix(contentType, "application/json") {
		body, err := ioutil.ReadAll(ctx.Request().Body)
		if err != nil {
			log.Errorf("authFilter 读取body数据异常:%+v", err)
			return "", err
		}
		ctx.Request().Body.Close()
		sign = util.Sign(app.Id, app.Secret, params, string(body))
		// request body被读取后需要进行恢复处理，否则后续逻辑读取不到body数据
		ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	} else {
		sign = util.Sign(app.Id, app.Secret, params)
	}
	return sign, nil
}

// 合并请求参数
func (s *AuthFilter) mergeRequestParams(myParams map[string]string, reqParams url.Values) {
	for k, v := range reqParams {
		// sign参数不参与签名计算
		if k == "sign" {
			continue
		}
		sort.Strings(v)
		myParams[k] = strings.Join(v, ",")
	}
}

// 获取请求参数
func (s *AuthFilter) getRequestParam(ctx web.Context, key string) string {
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

type AppInfo struct {
	// 应用ID
	Id int32 `json:"id"`
	// 密钥
	Secret string `json:"secret"`
	// 应用名称
	Name string `json:"name"`
}

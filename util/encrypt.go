package util

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"sort"

	"github.com/go-spring/spring-base/cast"
	"github.com/labstack/gommon/log"
)

// 接口签名生成
func Sign(appId int32, secret string, biz map[string]string, requestBody ...string) string {
	if biz == nil {
		biz = make(map[string]string)
	}
	// appId参数加入到业务参数map对象
	biz["appId"] = cast.ToString(appId)
	//取出map所有key进行正序排序
	keys := make([]string, 0, len(biz))
	for k := range biz {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// 按顺序进行拼接字符串
	builder := new(bytes.Buffer)
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString(biz[k])
	}
	// 如果content-type是application/json,则需要把请求体(@requestBody)进行拼接,注意：如果@requestBody为null则不需要进行拼接
	if len(requestBody) > 0 {
		builder.WriteString(requestBody[0])
	}
	// 最后拼接secretCode
	builder.WriteString(secret)
	// 生成md5
	h := md5.New()
	h.Write(builder.Bytes())
	signData := hex.EncodeToString(h.Sum(nil))
	log.Infof("util/Sign 明文参数:%s,生成签名:%s", builder.String(), signData)
	return signData
}

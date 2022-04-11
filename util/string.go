//Author:zhongzhenyu

package util

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"github.com/sony/sonyflake"
)

func RandString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func IsEmail(email string) bool {
	pattern := `^[0-9a-z][_.0-9a-z-]{0,31}@([0-9a-z][0-9a-z-]{0,30}[0-9a-z]\.){1,4}[a-z]{2,4}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}

func IsPhoneNo(mobileNum string) bool {
	regular := "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}

//转成字符串(变量.(断言类型))
func InterfaceToString(s interface{}) string {
	switch s.(type) {
	case float32:
		return strconv.FormatFloat(float64(s.(float32)), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(s.(float64), 'f', -1, 64)
	case int:
		return strconv.Itoa(s.(int))
	case int16:
		return strconv.FormatInt(int64(s.(int16)), 10)
	case int32:
		return strconv.FormatInt(int64(s.(int32)), 10)
	case int64:
		return strconv.FormatInt(s.(int64), 10)
	case uint8:
		return strconv.Itoa(int(s.(uint8)))
	case uint16:
		return strconv.Itoa(int(s.(uint16)))
	case uint32:
		return strconv.Itoa(int(s.(uint32)))
	case uint64:
		return strconv.Itoa(int(s.(uint64)))
	case string:
		return s.(string)
	default:
		panic(fmt.Sprintf("不支持将 %T 类型转成字符串", s))
	}
}

func IsEmptyString(s string) bool {
	if len(s) == 0 {
		return true
	}
	if len(strings.TrimSpace(s)) == 0 {
		return true
	}
	return false
}

//生成16位唯一字符编码
func GenSonyflake() string {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()

	if err == nil {
		return fmt.Sprintf("b%x", id)
	}
	return ""
}

//将Arry数据转换成map对象
func GenerateArryToMap(arry []string) map[string]interface{} {
	mapArry := make(map[string]interface{}, 0)
	for _, v := range arry {
		mapArry[v] = v
	}
	return mapArry
}

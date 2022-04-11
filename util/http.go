package util

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"github.com/go-spring/spring-core/web"
	"github.com/maybgit/glog"
	"github.com/ppkg/stark/dto"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func ResponseJSON(ctx web.Context, data dto.HttpResponse) {
	ctx.SetContentType(web.MIMEApplicationJSONCharsetUTF8)
	ctx.SetStatus(int(data.StatusCode()))
	ctx.JSON(data)
}
func HttpGet(url, source, ua string, param map[string]interface{}) (int, string) {
	return action(url, source, ua, http.MethodGet, "", param)
}

func HttpPostForm(url, source, ua string, param map[string]interface{}) (int, string) {
	return action(url, source, ua, http.MethodPost, "application/x-www-form-urlencoded", param)
}

func HttpPostJSON(url, source, ua string, param map[string]interface{}) (int, string) {
	return action(url, source, ua, http.MethodPost, "application/json;charset=UTF-8", param)
}

func action(uri, source, ua string, httpMethod string, contentType string, param map[string]interface{}) (int, string) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("http.action", err)
		}
	}()
	var req *http.Request
	switch httpMethod {
	case http.MethodGet:
		if param != nil {
			uri += "?" + mapToValues(param).Encode()
		}
		req, _ = http.NewRequest(httpMethod, uri, nil)
	case http.MethodPost:
		httpMethod = http.MethodPost
		var reader io.Reader

		if contentType == "application/x-www-form-urlencoded" {
			reader = strings.NewReader(mapToValues(param).Encode())
		} else if contentType == "application/json;charset=UTF-8" {
			byteData, _ := json.Marshal(param)
			reader = bytes.NewReader(byteData)
		}
		req, _ = http.NewRequest(httpMethod, uri, reader)
		req.Header.Add("Content-Type", contentType)
	default:
		return 0, "不支持的请求类型"
	}

	// for k, v := range httpHeader {
	// 	req.Header.Add(k, v)
	// }
	//ul := uuid.NewV4()
	//sn := strings.ReplaceAll(ul.String(), "-", "")
	//req.Header.Add("sn", sn)
	//req.Header.Add("source", source)
	//req.Header.Add("ua", ua)
	//req.Header.Add("timestamp", strconv.Itoa(int(time.Now().Unix())))

	client := http.Client{Timeout: time.Second * 30, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	res, err := client.Do(req)
	if err != nil {
		glog.Error(err)
		return 0, err.Error()
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

func mapToValues(mp map[string]interface{}) url.Values {
	v := url.Values{}
	for key, val := range mp {
		switch val.(type) {
		case int:
			v.Add(key, strconv.Itoa(val.(int)))
		case int32:
			v.Add(key, strconv.Itoa(int(val.(int32))))
		case int64:
			v.Add(key, strconv.Itoa(int(val.(int64))))
		case float64:
			v.Add(key, strconv.FormatFloat(val.(float64), 'E', -1, 64))
		case float32:
			v.Add(key, strconv.FormatFloat(float64(val.(float32)), 'E', -1, 32))
		default:
			v.Add(key, val.(string))
		}
	}
	//glog.Info(v.Encode())
	return v
}

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func HttpGetUrl(url string) string {

	// 超时时间：5秒
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}
	}

	return result.String()
}

// 发送GET请求
// url：         请求地址
// response：    请求返回的内容
func HttpGetUrlOMS(url string) (string,error) {

	// 超时时间：5秒
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
	}

	return result.String(),nil
}
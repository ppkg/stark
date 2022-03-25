package app

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"

	"github.com/ppkg/stark"

	"github.com/go-spring/spring-base/cast"
	"github.com/go-spring/spring-base/log"
	"github.com/go-spring/spring-base/util"
	"github.com/maybgit/glog"
)

var (
	flagEnv = flag.String("env", "", "set exec environment eg: dev,test,prod")
)

var (
	appInstanceOnce    int32
	appInstanceOnceErr = errors.New("the same app type can only be registered once")
)

func appInstanceOnceValidate() error {
	ok := atomic.CompareAndSwapInt32(&appInstanceOnce, 0, 1)
	if !ok {
		return appInstanceOnceErr
	}
	return nil
}

// 安装日志组件
func setupLogger() {
	log.SetOutput(log.FuncOutput(func(level log.Level, msg *log.Message) {
		defer func() { msg.Reuse() }()
		logFn := glog.Infof
		if level >= log.ErrorLevel {
			logFn = glog.Errorf
		} else if level == log.WarnLevel {
			logFn = glog.Warningf
		}
		var buf bytes.Buffer
		for _, a := range msg.Args() {
			buf.WriteString(cast.ToString(a))
		}
		fileLine := util.Contract(fmt.Sprintf("%s:%d", msg.File(), msg.Line()), 48)
		logFn("[%s] %s\n", fileLine, buf.String())
	}))
}

func initApplication(application *stark.Application) error {
	// 显示应用版本
	showAppVersion(application)

	// 安装日志组件
	setupLogger()

	// 初始化运行环境
	initRuntimeEnv()

	// 8. setup vars
	// setup app vars
	err := setupCommonVars(application)
	if err != nil {
		return err
	}
	// setup user vars
	if application.SetupVars != nil {
		err = application.SetupVars()
		if err != nil {
			return fmt.Errorf("application.SetupVars err: %v", err)
		}
	}
	return nil
}

// 初始化运行环境
func initRuntimeEnv(application *stark.Application) {
	if application.Environment != "" {
		return
	}
	// 运行环境
	if *flagEnv != "" {
		application.Environment = *flagEnv
		return
	}
	application.Environment = os.Getenv("env")
}

// setupCommonVars setup application global vars.
func setupCommonVars(application *stark.Application) error {
	// var err error
	// 安装数据库等

	return nil
}

func showAppVersion(app *stark.Application) {
	var logo = `%20__%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20___%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%0A%2F%5C%20%5C%20%20%20%20%20%20%20%20%20%20%20%20%20%2F%5C_%20%5C%20%20%20%20%20%20%20%20%20%20%20%20%20__%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%0A%5C%20%5C%20%5C%2F'%5C%20%20%20%20%20%20%20__%5C%2F%2F%5C%20%5C%20%20%20%20__%20%20__%20%2F%5C_%5C%20%20%20%20%20___%20%20%20%20%20%20____%20%20%0A%20%5C%20%5C%20%2C%20%3C%20%20%20%20%20%2F'__%60%5C%5C%20%5C%20%5C%20%20%2F%5C%20%5C%2F%5C%20%5C%5C%2F%5C%20%5C%20%20%2F'%20_%20%60%5C%20%20%20%2F'%2C__%5C%20%0A%20%20%5C%20%5C%20%5C%5C%60%5C%20%20%2F%5C%20%20__%2F%20%5C_%5C%20%5C_%5C%20%5C%20%5C_%2F%20%7C%5C%20%5C%20%5C%20%2F%5C%20%5C%2F%5C%20%5C%20%2F%5C__%2C%20%60%5C%0A%20%20%20%5C%20%5C_%5C%20%5C_%5C%5C%20%5C____%5C%2F%5C____%5C%5C%20%5C___%2F%20%20%5C%20%5C_%5C%5C%20%5C_%5C%20%5C_%5C%5C%2F%5C____%2F%0A%20%20%20%20%5C%2F_%2F%5C%2F_%2F%20%5C%2F____%2F%5C%2F____%2F%20%5C%2F__%2F%20%20%20%20%5C%2F_%2F%20%5C%2F_%2F%5C%2F_%2F%20%5C%2F___%2F%20`
	var version = `[Major Version：%v Type：%v]`
	var remote = `┌───────────────────────────────────────────────────┐
│ [Gitee] https://github.com/ppkg/stark      │
│ [GitHub] https://github.com/kelvins-io/kelvins    │
└───────────────────────────────────────────────────┘`
	fmt.Println("based on")
	logoS, _ := url.QueryUnescape(logo)
	fmt.Println(logoS)
	fmt.Println("")
	fmt.Println(fmt.Sprintf(version, stark.Version, stark.AppTypeText[app.Type]))

	fmt.Println("")
	fmt.Println(remote)
	fmt.Println("Go Go Go ==>", app.Name)
}

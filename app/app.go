package app

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ppkg/stark"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/go-spring/spring-base/cast"
	"github.com/go-spring/spring-base/log"
	"github.com/go-spring/spring-base/util"
	"github.com/go-spring/spring-core/gs"
	SpringGoRedis "github.com/go-spring/spring-go-redis"
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
	// 验证应用数据
	err := validateApplication(application)
	if err != nil {
		return err
	}

	// 显示应用版本
	showAppVersion(application)

	// 安装日志组件
	setupLogger()

	// 初始化运行环境
	initRuntimeEnv(application)

	// 加载默认配置
	err = loadDefaultConfig()
	if err != nil {
		return err
	}
	// 加载用户配置
	if application.LoadConfig != nil {
		err = application.LoadConfig()
		if err != nil {
			return fmt.Errorf("application.LoadConfig err: %v", err)
		}
	}

	// 安装组件
	err = setupCommonVars(application)
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

// 加载默认配置
func loadDefaultConfig() error {
	return nil
}

func validateApplication(application *stark.Application) error {
	if application.Name == "" {
		return fmt.Errorf("应用名称不能为空")
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
	var err error
	// 安装数据库组件
	err = setupDatabase(application)
	if err != nil {
		return err
	}

	return nil
}

// 安装各种数据库组件
func setupDatabase(application *stark.Application) error {
	list := application.GetDbConns()
	if len(list) == 0 {
		return nil
	}

	var err error
	for _, v := range list {
		switch v.Type {
		case stark.DbTypeMyql:
			err = setupMysql(application, v)
		case stark.DbTypeRedis:
			err = setupRedis(v)
		}
		if err != nil {
			log.Errorf("安装%s数据库(%s,%s)异常:%+v", stark.DbTypeText[v.Type], v.Name, v.Url, err)
			return err
		}
	}

	return nil
}

// 安装mysql
func setupMysql(application *stark.Application, info stark.DbConnInfo) error {
	gormConf := &gorm.Config{}
	if application.IsDebug {
		gormConf.Logger = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(mysql.Open(info.Url), gormConf)
	if err != nil {
		log.Errorf("打开%s数据库异常:%+v", stark.DbTypeText[info.Type], err)
		return err
	}
	// 设置数据库连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Errorf("设置%s数据库连接池异常:%+v", stark.DbTypeText[info.Type], err)
		return err
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	var temp int
	maxIdleConn := 10
	if val, ok := info.Extras["maxIdleConn"]; ok {
		temp = val.(int)
		if temp > 0 {
			maxIdleConn = temp
		}
	}
	sqlDB.SetMaxIdleConns(maxIdleConn)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	maxOpenConn := 1000
	if val, ok := info.Extras["maxOpenConn"]; ok {
		temp = val.(int)
		if temp > 0 {
			maxOpenConn = temp
		}
	}
	sqlDB.SetMaxOpenConns(maxOpenConn)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	connMaxLifetime := time.Hour
	if val, ok := info.Extras["connMaxLifetime"]; ok {
		connMaxLifetime = val.(time.Duration)
	}
	sqlDB.SetConnMaxLifetime(connMaxLifetime)

	gs.Object(db).Name(info.Name).Destroy(func(db *gorm.DB) {
		err = sqlDB.Close()
		if err != nil {
			log.Errorf("关闭%s数据库异常:%+v", stark.DbTypeText[info.Type], err)
		}
	})
	return nil
}

// 安装redis
func setupRedis(info stark.DbConnInfo) error {
	redisUrl := strings.Split(info.Url, ":")
	if len(redisUrl) != 2 {
		return fmt.Errorf("Redis连接地址不正确:%s", info.Url)
	}
	gs.Property("redis.host", redisUrl[0])
	gs.Property("redis.port", redisUrl[1])

	if val, ok := info.Extras["password"]; ok {
		gs.Property("redis.password", val)
	}
	if val, ok := info.Extras["db"]; ok {
		gs.Property("redis.database", val)
	}
	gs.Provide(SpringGoRedis.NewClient, "${redis}").Name(info.Name)
	return nil
}

func showAppVersion(app *stark.Application) {
	var logo = `%20ad88888ba%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%2088%20%20%20%20%20%20%20%20%20%0Ad8%22%20%20%20%20%20%228b%20%20%2Cd%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%2088%20%20%20%20%20%20%20%20%20%0AY8%2C%20%20%20%20%20%20%20%20%20%2088%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20%2088%20%20%20%20%20%20%20%20%20%0A%60Y8aaaaa%2C%20%20MM88MMM%20%20%2CadPPYYba%2C%20%208b%2CdPPYba%2C%20%2088%20%20%20%2Cd8%20%20%20%0A%20%20%60%22%22%22%22%228b%2C%20%2088%20%20%20%20%20%22%22%20%20%20%20%20%60Y8%20%2088P'%20%20%20%22Y8%20%2088%20%2Ca8%22%20%20%20%20%0A%20%20%20%20%20%20%20%20%608b%20%2088%20%20%20%20%20%2CadPPPPP88%20%2088%20%20%20%20%20%20%20%20%20%208888%5B%20%20%20%20%20%20%0AY8a%20%20%20%20%20a8P%20%2088%2C%20%20%20%2088%2C%20%20%20%20%2C88%20%2088%20%20%20%20%20%20%20%20%20%2088%60%22Yba%2C%20%20%20%0A%20%22Y88888P%22%20%20%20%22Y888%20%20%60%228bbdP%22Y8%20%2088%20%20%20%20%20%20%20%20%20%2088%20%20%20%60Y8a%20%20`
	var version = `[Major Version：%v Type：%v]`
	logoS, _ := url.QueryUnescape(logo)
	fmt.Println(logoS)
	fmt.Println("")
	fmt.Println(fmt.Sprintf(version, stark.Version, stark.AppTypeText[app.Type]))
}

module github.com/ppkg/stark

go 1.16

require (
	github.com/go-spring/spring-base v1.1.0-rc3
	github.com/go-spring/spring-core v1.1.0-rc3
	github.com/go-spring/spring-go-redis v1.1.0-rc3
	github.com/go-spring/starter-echo v1.1.0-rc3
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/maybgit/glog v0.1.22
	github.com/ucarion/urlpath v0.0.0-20200424170820-7ccc79b76bbb
	google.golang.org/grpc v1.45.0
	gorm.io/driver/mysql v1.3.2
	gorm.io/gorm v1.23.3

)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.11.4
	github.com/kr/text v0.2.0 // indirect
	github.com/labstack/echo/v4 v4.6.1
	github.com/limitedlee/microservice v0.1.7
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/sony/sonyflake v1.0.0
	github.com/spf13/cast v1.4.1
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)

replace github.com/go-spring/spring-core v1.1.0-rc3 => github.com/ppkg/spring-core v1.2.4

// replace github.com/go-spring/spring-core => /home/zihua/Documents/goPath/src/github.com/go-spring/spring-core

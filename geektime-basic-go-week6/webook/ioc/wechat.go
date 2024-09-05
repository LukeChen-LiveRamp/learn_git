package ioc

import (
	"gitee.com/geekbang/basic-go/webook/internal/service/oauth2/wechat"
	"gitee.com/geekbang/basic-go/webook/pkg/logger"
	"os"
)

func InitWechatService(l logger.LoggerV1) wechat.Service {
	appID, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		//panic("找不到环境变量 WECHAT_APP_ID")
		appID = "wx123456"
	}
	appSecret, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		//panic("找不到环境变量 WECHAT_APP_SECRET")
		appSecret = "wx123456"
	}
	return wechat.NewService(appID, appSecret, l)
}

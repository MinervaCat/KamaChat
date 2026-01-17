package main

import (
	"fmt"
	"kama_chat_server/internal/config"
	"kama_chat_server/internal/https_server"
	"kama_chat_server/internal/service/kafka"
	"kama_chat_server/internal/service/push"
	"kama_chat_server/pkg/zlog"
)

func main() {
	zlog.Info("push服务开始")
	conf := config.GetConfig()
	host := conf.MainConfig.Host
	port := conf.MainConfig.Port
	go func() {
		// Ubuntu22.04云服务器部署
		if err := https_server.GE.RunTLS(fmt.Sprintf("%s:%d", host, port), "/etc/ssl/certs/server.crt", "/etc/ssl/private/server.key"); err != nil {
			zlog.Fatal("server running fault")
			return
		}
	}()
	kafka.KafkaService.KafkaInit2()
	push.Pusher.Start()
}

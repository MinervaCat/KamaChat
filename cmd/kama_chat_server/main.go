package main

import (
	"fmt"
	"kama_chat_server/internal/config"
	"kama_chat_server/internal/https_server"
	"kama_chat_server/internal/service/chat"
	"kama_chat_server/internal/service/kafka"
	myredis "kama_chat_server/internal/service/redis"
	"kama_chat_server/pkg/zlog"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.Println("服务开始")
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
	kafka.KafkaService.KafkaInit()

	go chat.ConversationServer.Start()
	go chat.UserServer.Start()

	// 设置信号监听
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 等待信号
	<-quit

	kafka.KafkaService.KafkaClose()

	zlog.Info("关闭服务器...")

	// 删除所有Redis键
	if err := myredis.DeleteAllRedisKeys(); err != nil {
		zlog.Error(err.Error())
	} else {
		zlog.Info("所有Redis键已删除")
	}

	zlog.Info("服务器已关闭")

}

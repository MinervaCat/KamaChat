package main

import (
	"kama_chat_server/grpc_server"
	"kama_chat_server/internal/service/chat"
	"kama_chat_server/internal/service/kafka"
	myredis "kama_chat_server/internal/service/redis"
	"kama_chat_server/pkg/zlog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	zlog.Info("chat_server服务开始")

	kafka.KafkaService.KafkaInit()

	go chat.ConversationServer.Start()
	go chat.UserServer.Start()
	go grpc_server.Server.Start()
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

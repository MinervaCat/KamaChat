package push_server

import (
	"kama_chat_server/internal/service/kafka"
	"kama_chat_server/internal/service/push"
)

func main() {
	kafka.KafkaService.KafkaInit2()
	push.Pusher.Start()
}

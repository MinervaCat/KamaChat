package push_server

import "kama_chat_server/internal/service/push"

func main() {
	push.Pusher.Start()
}

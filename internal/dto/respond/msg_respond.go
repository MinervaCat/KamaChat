package respond

import "time"

type MessageRespond struct {
	MsgId          int64     `json:"msg_id"`
	ConversationId string    `json:"conversation_id"`
	Seq            int64     `json:"seq"`
	SendId         int64     `json:"send_id"`
	Type           int8      `json:"type"`
	Content        string    `json:"content"`
	Status         int8      `json:"status"`
	SendTime       time.Time `json:"send_time"`
}

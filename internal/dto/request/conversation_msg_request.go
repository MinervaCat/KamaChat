package request

import "time"

type ConversationMsgRequest struct {
	ConversationId string    `json:"conversation_id"`
	Type           int8      `json:"type"`
	Content        string    `json:"content"`
	UserId         int64     `json:"user_id"`
	MsgId          int64     `json:"msg_id"`
	SendTime       time.Time `json:"send_time"`
}

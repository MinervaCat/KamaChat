package request

type UserMsgRequest struct {
	UserId         int64  `json:"user_id"`
	MsgId          int64  `json:"msg_id"`
	ConversationId string `json:"conversation_id"`
}

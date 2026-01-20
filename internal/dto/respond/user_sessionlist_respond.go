package respond

type ConversationListRespond struct {
	ConversationId string `json:"conversation_id"`
	LastReadSeq    int64  `json:"last_read_seq"`
	NotifyType     int8   `json:"notify_type"`
	IsTop          int8   `json:"is_top"`
}

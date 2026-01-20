package respond

import "time"

type ConversationListRespond struct {
	ConversationId string    `json:"conversation_id"`
	Avatar         string    `json:"avatar"`
	Type           int8      `json:"type"`
	Member         int32     `json:"member"`
	RecentMsgTime  time.Time `json:"recent_msg_time"`
	LastReadSeq    int64     `json:"last_read_seq"`
	NotifyType     int8      `json:"notify_type"`
	IsTop          int8      `json:"is_top"`
}

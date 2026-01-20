package model

type UserConversationList struct {
	Id             int64  `gorm:"column:id;primaryKey;comment:自增id"`
	UserId         int64  `gorm:"column:user_id;index:conversation_user_idx,priority:2;not null"`
	ConversationId string `gorm:"column:conversation_id;index:conversation_user_idx,priority:1;type:varchar(64);comment:会话ID"`
	LastReadSeq    int64  `gorm:"column:last_read_seq"`
	NotifyType     int8   `gorm:"column:notify_type"`
	IsTop          int8   `gorm:"column:is_top"`
}

// todo user_conversation_idx
func (UserConversationList) TableName() string {
	return "user_conversation_list"
}

package model

type UserMsgList struct {
	Id             int64  `gorm:"column:id;primaryKey;comment:自增id" json:"user_id"`
	UserId         int64  `gorm:"column:user_id;not null;index:sort_msg_idx,priority:1"`
	MsgId          int64  `gorm:"column:msg_id;not null"`
	ConversationId string `gorm:"column:conversation_id;type:varchar(64);comment:会话ID"`
	Seq            int64  `gorm:"column:seq;index:sort_msg_idx,priority:2;not null"`
}

func (UserMsgList) TableName() string {
	return "user_msg_list"
}

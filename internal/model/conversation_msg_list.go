package model

type ConversationMsgList struct {
	Id             int64  `gorm:"column:id;primaryKey;comment:自增id"`
	ConversationId string `gorm:"column:conversation_id;index:sort_msg_idx,priority:1;type:varchar(64);comment:会话ID"`
	MsgId          int64  `gorm:"column:msg_id;not null"`
	Seq            int64  `gorm:"column:seq;index:sort_msg_idx,priority:2;not null"`
	//DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;type:datetime;comment:删除时间"`
}

func (ConversationMsgList) TableName() string {
	return "conversation_msg_list"
}

package model

import (
	"gorm.io/gorm"
	"time"
)

type Conversation struct {
	Id             int64          `gorm:"column:id;primaryKey;comment:自增id"`
	ConversationId string         `gorm:"column:conversation_id;uniqueIndex;type:varchar(64);comment:会话ID"`
	Type           int8           `gorm:"column:type;not null;comment:会话类型，0单聊，1群聊"`
	Member         int32          `gorm:"column:member;not null"`
	Avatar         string         `gorm:"column:avatar;type:varchar(128);comment:表示群组头像"`
	RecentMsgTime  time.Time      `gorm:"column:recent_msg_time;type:datetime;not null;comment:此会话最新消息时间"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index;type:datetime;comment:删除时间"`
}

func (Conversation) TableName() string {
	return "conversation"
}

package model

import (
	"time"
)

type Msg struct {
	Id             int64     `gorm:"column:id;primaryKey;comment:自增id"`
	MsgId          int64     `gorm:"column:msg_id;uniqueIndex;not null;comment:消息uuid"`
	ConversationId string    `gorm:"column:conversation_id;type:varchar(64);comment:会话ID"`
	Type           int8      `gorm:"column:type;not null;comment:消息类型，0.文本，1.语音，2.文件，3.通话"` // 通话不用存消息内容或者url
	Content        string    `gorm:"column:content;type:TEXT;comment:消息内容"`
	Url            string    `gorm:"column:url;type:char(255);comment:消息url"`
	UserId         int64     `gorm:"column:user_id;not null;comment:发送者uuid"`
	FileType       string    `gorm:"column:file_type;type:char(10);comment:文件类型"`
	FileName       string    `gorm:"column:file_name;type:varchar(50);comment:文件名"`
	FileSize       string    `gorm:"column:file_size;type:char(20);comment:文件大小"`
	Status         int8      `gorm:"column:status;not null;comment:状态，0.未发送，1.已发送"`
	SendTime       time.Time `gorm:"column:send_time;type:datetime;comment:发送时间"`
	AVData         string    `gorm:"column:av_data;comment:通话传递数据"`
}

func (Msg) TableName() string {
	return "msg"
}

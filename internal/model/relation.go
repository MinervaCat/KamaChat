package model

type Relation struct {
	Id       int64 `gorm:"column:id;primaryKey;comment:自增id"`
	UserId   int64 `gorm:"column:user_id;uniqueIndex:user_friend_idx,priority:1;not null"`
	FriendId int64 `gorm:"column:friend_id;uniqueIndex:user_friend_idx,priority:2;not null"`
	Status   int8  `gorm:"column:status;not null;comment:状态，0.正常，1.拉黑，2.被拉黑，3.删除好友，4.被删除好友"`
}

func (Relation) TableName() string {
	return "relation"
}

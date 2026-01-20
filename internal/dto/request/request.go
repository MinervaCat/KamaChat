package request

type UserRequest struct {
	UserId int64 `json:"user_id"`
}

type UserFriendRequest struct {
	UserId   int64 `json:"user_id"`
	FriendId int64 `json:"friend_id"`
}

type ConversationRequest struct {
	ConversationId string `json:"conversation_id"`
}

type UserConversationRequest struct {
	UserId         int64  `json:"user_id"`
	ConversationId string `json:"conversation_id"`
}

type UserSeqRequest struct {
	UserId int64 `json:"user_id"`
	Seq    int64 `json:"seq"`
}

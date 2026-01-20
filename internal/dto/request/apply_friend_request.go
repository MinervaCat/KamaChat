package request

type ApplyFriendRequest struct {
	UserId   int64  `json:"user_id"`
	FriendId int64  `json:"friend_id"`
	Message  string `json:"message"`
}

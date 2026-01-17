package respond

type GetUserListRespond struct {
	UserId   int64  `json:"user_id"`
	Nickname string `json:"nickname"`

	Status    int8 `json:"status"`
	IsAdmin   int8 `json:"is_admin"`
	IsDeleted bool `json:"is_deleted"`
}

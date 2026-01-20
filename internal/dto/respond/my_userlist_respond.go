package respond

type MyUserListRespond struct {
	UserId   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	Avatar   string `json:"avatar"`
}

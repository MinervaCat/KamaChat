package respond

type GetUserInfoRespond struct {
	UserId    int64  `json:"user_id"`
	Nickname  string `json:"nickname"`
	Telephone string `json:"telephone"`
	Avatar    string `json:"avatar"`
	Email     string `json:"email"`
	Gender    int8   `json:"gender"`
	Birthday  string `json:"birthday"`
	Signature string `json:"signature"`
	CreatedAt string `json:"created_at"`
	IsAdmin   int8   `json:"is_admin"`
	Status    int8   `json:"status"`
}

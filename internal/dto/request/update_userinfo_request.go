package request

type UpdateUserInfoRequest struct {
	UserId    int64  `json:"user_id"`
	Email     string `json:"email"`
	Nickname  string `json:"nickname"`
	Birthday  string `json:"birthday"`
	Signature string `json:"signature"`
	Avatar    string `json:"avatar"`
}

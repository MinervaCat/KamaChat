package request

type LoginRequest struct {
	Telephone int64  `json:"telephone"`
	Password  string `json:"password"`
}

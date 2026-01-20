package request

type RegisterRequest struct {
	Telephone string `json:"telephone"`
	Password  string `json:"password"`
	Nickname  string `json:"nickname"`
	SmsCode   string `json:"sms_code"`
}

type RegisterRequest2 struct {
	Telephone string `json:"telephone"`
	Password  string `json:"password"`
	Nickname  string `json:"nickname"`
}

package respond

type NewContactListRespond struct {
	ContactId     int64  `json:"contact_id"`
	ContactName   string `json:"contact_name"`
	ContactAvatar string `json:"contact_avatar"`
	Message       string `json:"message"`
}

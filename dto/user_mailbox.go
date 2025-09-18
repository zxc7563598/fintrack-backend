package dto

type UserMailboxListItem struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	AuthCode string `json:"auth_code"`
	IMAP     string `json:"imap"`
	Remark   string `json:"remark"`
}

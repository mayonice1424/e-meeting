package models

type User struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	No_HP           string `json:"no_hp"`
	Role            string `json:"role"`
	Status          string `json:"status"`
	Language        string `json:"language"`
	Profile_Picture string `json:"profile_picture"`
	Confirm_Password string `json:"confirm_password"`
}
type ErrorResponse struct {
	Message string `json:"message"`
}
type UserId struct {
	Data    []User                 `json:"data"`
	Message map[string]interface{} `json:"message,omitempty"`
}
type SuccessResponse struct {
	Data    Token                 `json:"data"`
	Message string `json:"message"`

}
type Token struct {
	Token string `json:"token"`
}

package models


type User struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Email            string `json:"email"`
	Username         string `json:"username"`
	No_HP            string `json:"no_hp"`
	Role             string `json:"role"`
	Status           string `json:"status"`
	Language         string `json:"language"`
	Profile_Picture  string `json:"profile_picture"`
	Created_At       string `json:"created_at"`
	Updated_At       string `json:"updated_at"`
}

type CreateUser struct {
	Email            string `json:"email"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	Confirm_Password string `json:"confirm_password"`
}

type LoginUser struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
}
type ErrorResponse struct {
	Message string `json:"message"`
}
type UserId struct {
	Data    []User                 `json:"data"`
	Message map[string]interface{} `json:"message,omitempty"`
}
type SuccessResponse struct {
	Data    *int   `json:"data"`
	Message string `json:"message"`
}

type SuccessResponseLogin struct {
	Data    Token   `json:"data"`
	Message string `json:"message"`
}
type Token struct {
	Access_Token  string `json:"accessToken"`
	Refresh_Token string `json:"refreshToken"`
}

// type PasswordReset struct {
// 	ID        int       `json:"id"`
// 	UserID    int       `json:"user_id"`
// 	Token     string    `json:"token"`
// 	ExpiresAt time.Time `json:"expires_at"`
// 	CreatedAt time.Time `json:"created_at"`
// }

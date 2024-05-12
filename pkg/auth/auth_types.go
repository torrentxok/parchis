package auth

type User struct {
	User_id    int    `json:"user_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	UserGroup  string `json:"user_group"`
	IsVerified bool   `json:"is_verified"`
	Isdeleted  int    `json:"is_deleted"`
	Password   string `json:"password"`
}

type RegistrationData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	UserId       int    `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	UserGroup    string `json:"user_group"`
	IsVerified   bool   `json:"is_verified"`
	Isdeleted    int    `json:"is_deleted"`
	Password     string `json:"password"`
	PasswordHash []byte `json:"password_hash"`
}

type RegistrationData struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserSession struct {
	SessionId    uuid.UUID `json:"session_id"`
	UserId       int       `json:"user_id"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	CreationDate time.Time `json:"creation_date"`
	UpdateDate   time.Time `json:"updated_date"`
	ExpiryTime   time.Time `json:"expiry_time"`
}

type JSONResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

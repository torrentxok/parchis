package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/torrentxok/parchis/pkg/cfg"
	database "github.com/torrentxok/parchis/pkg/db"
)

// Валидация данных при регистрации
func ValidateRegData(regData RegistrationData) error {
	if regData.Email == "" || regData.Username == "" || regData.Password == "" {
		return errors.New("имя пользователя, адрес электронной почты и пароль обязательны")
	}

	return nil
}

func ValidateLoginData(login LoginData) error {
	if login.Email == "" || login.Password == "" {
		return errors.New("почта и пароль обязательны")
	}
	return nil
}

// Добавление пользователя в БД (подумать как сделать, в виде ХП)
func InsertUser(regData RegistrationData) (*User, error) {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return nil, err
	}
	log.Print("[INFO] Connected to DB")
	defer db.Close(context.Background())
	u := User{
		UserId:     -1,
		Username:   regData.Username,
		Email:      regData.Email,
		UserGroup:  "",
		IsVerified: false,
		Isdeleted:  0,
		Password:   "",
	}
	_password, err := HashPassword(regData.Password)
	if err != nil {
		log.Print("[ERROR] Password hash generation error: " + err.Error())
		return nil, err
	}
	u.Password = _password
	log.Print("[INFO] Password hash generated")

	err = insertUserToDB(db, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func LoginUser(login *LoginData) (string, string, error) {
	log.Print("[INFO] Start login user")
	var u User
	u.Email = login.Email
	u.Password = login.Password

	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return "", "", err
	}
	err = getUserData(db, &u)
	if err != nil {
		log.Print("[ERROR] Ошибка поиска пользователя: " + err.Error())
		return "", "", err
	}

	err = bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(login.Password))
	if err != nil {
		log.Print("[ERROR] Неверный пароль: " + err.Error())
		return "", "", err
	}

	if !u.IsVerified {
		log.Print("[ERROR] Пользователь не подтвержден")
		return "", "", errors.New("[ERROR] Пользователь не подтвержден")
	}

	session, err := CreateSession(&u)
	if err != nil {
		log.Print("[ERROR] Ошибка создания сессии: " + err.Error())
		return "", "", err
	}
	return session.AccessToken, session.RefreshToken, nil
}

func CreateSession(u *User) (*UserSession, error) {
	us := UserSession{
		SessionId:    uuid.New(),
		UserId:       u.UserId,
		CreationDate: time.Now(),
		UpdateDate:   time.Now(),
		ExpiryTime:   time.Now().Add(24 * time.Hour),
	}
	var err error
	us.AccessToken, err = GenerateToken(u, time.Minute*15)
	if err != nil {
		return nil, err
	}
	us.RefreshToken, err = GenerateToken(u, time.Hour*24)
	if err != nil {
		return nil, err
	}
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return nil, err
	}
	err = insertUserSession(db, &us)
	if err != nil {
		log.Print("[ERROR] Ошибка добавления сессии: " + err.Error())
		return nil, err
	}
	return &us, nil
}

func GenerateToken(u *User, t time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["UserId"] = u.UserId
	claims["exp"] = time.Now().Add(t).Unix()
	tokenString, err := token.SignedString(cfg.JWTKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Генерация хеша пароля
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// Отправка сообщения подтверждения
func SendConfirmationEmail(u *User) error {
	log.Print("[INFO] Start send email")

	confirmationToken, err := generateHashedToken(u)
	if err != nil {
		log.Print("Ошибка при генерации токена" + err.Error())
		return err
	}

	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	log.Print("[INFO] Connected to DB")
	defer db.Close(context.Background())

	err = insertConfirmationTokenToDB(db, u, confirmationToken)
	if err != nil {
		return err
	}

	confirmationUrl := fmt.Sprintf("Для подтверждения почты перейдите по ссылке: localhost:8080/confirm_email?token=%s", confirmationToken)

	err = SendEmail(u.Email, confirmationUrl)
	if err != nil {
		return err
	}
	return nil
}

func SendEmail(email string, text string) error {
	auth := smtp.PlainAuth("", cfg.ConfigVar.SMTP.SenderEmail, cfg.ConfigVar.SMTP.SenderPasswd, "smtp.gmail.com")
	to := []string{email}

	msg := []byte("To:" + email + "\r\n" +
		"Subject: Parchis: Confirm Email\r\n" +
		"\r\n" +
		text + "\r\n")

	err := smtp.SendMail("smtp.gmail.com:25", auth, cfg.ConfigVar.SMTP.SenderEmail, to, msg)
	if err != nil {
		return err
	}
	return nil
}

func generateHashedToken(u *User) (string, error) {
	// Генерируем случайные данные
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Создаем строку из имени пользователя, текущего времени и случайных данных
	token := fmt.Sprintf("%s:%d:%x", u.Email, time.Now().Unix(), b)

	hash := sha256.Sum256([]byte(token))

	// Генерируем хеш от этой строки с использованием bcrypt
	hashedToken, err := bcrypt.GenerateFromPassword(hash[:], bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedToken), nil
}

func ConfirmEmail(token string) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	log.Print("[INFO] Connected to DB")
	defer db.Close(context.Background())

	err = confirmEmailInDB(db, token)
	if err != nil {
		return err
	}

	log.Print("[INFO] Email confirmed")

	return nil
}

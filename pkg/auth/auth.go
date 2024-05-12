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
		User_id:    -1,
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

// Генерация хеша пароля
func HashPassword(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	hashedPassword, err := bcrypt.GenerateFromPassword(hash[:], bcrypt.DefaultCost)
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

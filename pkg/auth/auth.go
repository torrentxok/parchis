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

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/torrentxok/parchis/pkg/cfg"
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
	db, err := ConnectToDB()
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

	var DBerror pgtype.Int4
	err = db.QueryRow(context.Background(),
		`SELECT user_id, error_msg FROM dbo.insert_user(
			username => $1,
			email => $2,
			password_hash => $3)`,
		u.Username, u.Email, u.Password).Scan(&u.User_id, &DBerror)
	if err != nil {
		log.Print("[ERROR] Error insert user", err.Error())
		return nil, err
	}
	log.Print(u.User_id, DBerror.Int, DBerror.Status)
	if DBerror.Status == pgtype.Present && u.User_id == -1 {
		switch DBerror.Int {
		case 1:
			return nil, errors.New("[ERROR] Пользователь уже существует")
		default:
			return nil, errors.New("[ERROR] Непредвиденная ошибка")
		}
	} else {
		log.Print(u.User_id)
	}

	return &u, nil
}

// Подключение к БД
func ConnectToDB() (*pgx.Conn, error) {
	dbConnectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", cfg.ConfigVar.Database.User,
		cfg.ConfigVar.Database.Password, cfg.ConfigVar.Database.DBName)
	db, err := pgx.Connect(context.Background(), dbConnectionString)
	if err != nil {
		log.Print("[ERROR] Error connection to DB: ", err)
		return nil, err
	}
	return db, nil
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

	db, err := ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	log.Print("[INFO] Connected to DB")
	var DBerror pgtype.Int4
	err = db.QueryRow(context.Background(),
		`SELECT * FROM dbo.insert_confirmation_token(
			_user_id => $1,
			_email => $2,
			confirmation_token => $3)`,
		u.User_id, u.Email, confirmationToken).Scan(&DBerror)
	if err != nil {
		log.Print("[ERROR] Error insert token" + err.Error())
		return err
	}

	if DBerror.Status == pgtype.Present {
		switch DBerror.Int {
		case 1:
			return errors.New("[ERROR] Пользователя нет в системе")
		default:
			return errors.New("[ERROR] Непредвиденная ошибка")
		}
	}

	confirmationUrl := fmt.Sprintf("localhost:8080/confirm_email?token=%s", confirmationToken)

	auth := smtp.PlainAuth("", cfg.ConfigVar.SMTP.SenderEmail, cfg.ConfigVar.SMTP.SenderPasswd, "smtp.gmail.com")
	to := []string{u.Email}

	msg := []byte("To:" + u.Email + "\r\n" +
		"Subject: Parchis: Confirm Email\r\n" +
		"\r\n" +
		"Перейдите по ссылке для подтверждения почты: " + confirmationUrl + "\r\n")

	err = smtp.SendMail("smtp.gmail.com:25", auth, cfg.ConfigVar.SMTP.SenderEmail, to, msg)
	if err != nil {
		log.Fatal(err.Error())
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
	db, err := ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	log.Print("[INFO] Connected to DB")

	var DBerror pgtype.Int4
	err = db.QueryRow(context.Background(),
		`SELECT * FROM dbo.confirm_email(
			_confirmation_token => $1)`,
		token).Scan(&DBerror)
	if err != nil {
		log.Print("[ERROR] Error confirm email: " + err.Error())
		return err
	}

	if DBerror.Status == pgtype.Present {
		switch DBerror.Int {
		case 1:
			return errors.New("[ERROR] Пользователь не найден")
		default:
			return errors.New("[ERROR] Непредвиденная ошибка")
		}
	}

	log.Print("[INFO] Email confirmed")

	return nil
}

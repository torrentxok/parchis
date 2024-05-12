package auth

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
	"github.com/torrentxok/parchis/pkg/cfg"
)

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

func insertUserToDB(db *pgx.Conn, u *User) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT user_id, error_msg FROM dbo.insert_user(
			username => $1,
			email => $2,
			password_hash => $3)`,
		u.Username, u.Email, u.Password).Scan(&u.User_id, &DBerror)
	if err != nil {
		log.Print("[ERROR] Error insert user", err.Error())
		return err
	}
	if DBerror.Status == pgtype.Present && u.User_id == -1 {
		switch DBerror.Int {
		case 1:
			return errors.New("[ERROR] Пользователь уже существует")
		default:
			return errors.New("[ERROR] Непредвиденная ошибка")
		}
	} else {
		log.Print("[INFO] Пользователь добавлен")
	}

	return nil
}

func insertConfirmationTokenToDB(db *pgx.Conn, u *User, confirmationToken string) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
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
	return nil
}

func confirmEmailInDB(db *pgx.Conn, token string) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.confirm_email(
			_confirmation_token => $1)`,
		token).Scan(&DBerror)
	if err != nil {
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
	return nil
}

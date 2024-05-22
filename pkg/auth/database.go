package auth

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
)

func insertUserToDB(db *pgx.Conn, u *User) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT user_id, error_msg FROM dbo.insert_user(
			username => $1,
			email => $2,
			password_hash => $3)`,
		u.Username, u.Email, u.Password).Scan(&u.UserId, &DBerror)
	if err != nil {
		log.Print("[ERROR] Error insert user", err.Error())
		return err
	}
	if DBerror.Status == pgtype.Present && u.UserId == -1 {
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
		u.UserId, u.Email, confirmationToken).Scan(&DBerror)
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

func getUserData(db *pgx.Conn, u *User) error {
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.search_user(
			_email => $1)`,
		u.Email).
		Scan(&u.UserId, &u.Username, &u.Email, &u.UserGroup, &u.IsVerified, &u.Isdeleted, &u.PasswordHash)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return errors.New("[ERROR] Пользователь не найден")
		default:
			return err
		}
	}
	return nil
}

func insertUserSession(db *pgx.Conn, us *UserSession) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.insert_user_session(
			_session_id => $1,
			_user_id => $2,
			_access_token => $3,
			_refresh_token => $4,
			_creation_date => $5,
			_updated_date => $6,
			_access_token_expiry_time => $7,
			_refresh_token_expiry_time => $8)`,
		us.SessionId, us.UserId, us.AccessToken, us.RefreshToken, us.CreationDate, us.UpdateDate, us.AccessTokenExpiryTime, us.RefreshTokenExpiryTime).
		Scan(&DBerror)
	if err != nil {
		return err
	}
	return nil
}

func validateTokenInDB(db *pgx.Conn, token string) (bool, error) {
	var ok bool
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.validate_access_token(
			p_access_token => $1)`,
		token).Scan(&ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func getUserSession(db *pgx.Conn, userId int, refreshToken string) (UserSession, error) {
	var us UserSession
	var tempAccess []byte
	var tempRefresh []byte
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.get_user_session(
			p_user_id => $1,
			p_refresh_token => $2)`,
		userId, refreshToken).
		Scan(&us.SessionId, &us.UserId, &tempAccess, &tempRefresh, &us.CreationDate,
			&us.UpdateDate, &us.AccessTokenExpiryTime, &us.RefreshTokenExpiryTime, &us.EndDate)
	if err != nil {
		return us, err
	}
	us.AccessToken = string(tempAccess)
	us.RefreshToken = string(tempRefresh)
	return us, nil
}

func updateUserSession(db *pgx.Conn, us *UserSession) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.update_user_session(
			p_session_id => $1,
			p_user_id => $2,
			p_access_token => $3,
			p_refresh_token => $4,
			p_updated_date => $5,
			p_access_token_expiry_time => $6,
			p_refresh_token_expiry_time => $7)`,
		us.SessionId, us.UserId, us.AccessToken, us.RefreshToken,
		us.UpdateDate, us.AccessTokenExpiryTime, us.RefreshTokenExpiryTime).
		Scan(&DBerror)
	if err != nil {
		return err
	}
	return nil
}

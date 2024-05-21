package user

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
)

func GetUserProfileInfo(db *pgx.Conn, targetUserId int, currentUserId int) (UserProfileResponse, error) {
	var profile UserProfileResponse
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.get_user_profile_info(
			target_user_id => $1,
			current_user_id => $2)`,
		targetUserId, currentUserId).Scan(&profile.Profile.Id, &profile.Profile.Username, &profile.IsOwnProfile, &profile.FriendshipStatus)
	if err != nil {
		log.Print("[ERROR] Error get user info")
		return profile, err
	}
	return profile, nil
}

func GetUserFriends(db *pgx.Conn, targetUserId int) ([]struct {
	Id       int    `json:"id"`
	Username string `json:"username"`
}, error) {

	var friends []struct {
		Id       int    `json:"id"`
		Username string `json:"username"`
	}

	rows, err := db.Query(context.Background(),
		`SELECT friend_id, friend_username FROM dbo.get_user_friends(
			target_user_id => $1)`, targetUserId)
	if err != nil {
		log.Print("[ERROR] Error get user friends")
		return friends, err
	}
	defer rows.Close()

	for rows.Next() {
		var friend struct {
			Id       int    `json:"id"`
			Username string `json:"username"`
		}
		err := rows.Scan(&friend.Id, &friend.Username)
		if err != nil {
			return friends, err
		}
		friends = append(friends, friend)
	}

	return friends, nil
}

func FriendshipsManaging(db *pgx.Conn, targetUserId int, currentUserId int, action string) error {
	var DBerror pgtype.Int4
	switch action {
	case "request":
		err := db.QueryRow(context.Background(),
			`SELECT * FROM dbo.friendship_request(
				target_user_id => $1,
				current_user_id => $2)`,
			targetUserId, currentUserId).Scan(&DBerror)
		if err != nil {
			return err
		}
		if DBerror.Status == pgtype.Present {
			switch DBerror.Int {
			case 1:
				return errors.New("[ERROR] Одного или нескольких пользователей не существует")
			default:
				return errors.New("[ERROR] Непредвиденная ошибка")
			}
		}
	case "accept":
		err := db.QueryRow(context.Background(),
			`SELECT * FROM dbo.friendship_accept(
				target_user_id => $1,
				current_user_id => $2)`,
			targetUserId, currentUserId).Scan(&DBerror)
		if err != nil {
			return err
		}
		if DBerror.Status == pgtype.Present {
			switch DBerror.Int {
			case 1:
				return errors.New("[ERROR] Записи не существует")
			default:
				return errors.New("[ERROR] Непредвиденная ошибка")
			}
		}
	case "remove":
		err := db.QueryRow(context.Background(),
			`SELECT * FROM dbo.friendship_remove(
				target_user_id => $1,
				current_user_id => $2)`,
			targetUserId, currentUserId).Scan(&DBerror)
		if err != nil {
			return err
		}
		if DBerror.Status == pgtype.Present {
			switch DBerror.Int {
			case 1:
				return errors.New("[ERROR] Записи не существует")
			default:
				return errors.New("[ERROR] Непредвиденная ошибка")
			}
		}
	default:
		log.Print("[ERROR] Wrong action")
		return errors.New("wrong action")
	}
	return nil
}

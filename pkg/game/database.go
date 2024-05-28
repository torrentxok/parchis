package game

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
)

func IsPlayerInGameDB(db *pgx.Conn, gameId int, userId int) (bool, error) {
	var inGame bool
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.is_player_in_game(
			p_game_id => $1,
			p_user_id => $2)`,
		gameId, userId).
		Scan(inGame)
	if err != nil {
		return false, err
	}
	return inGame, nil
}

func ExistsGame(db *pgx.Conn, gameId int) (bool, error) {
	var existsGame bool
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.exists_game(
			p_game_id => $1)`,
		gameId).
		Scan(&existsGame)
	if err != nil {
		return false, err
	}
	return existsGame, nil
}

func AddNewStateToDB(db *pgx.Conn, bs BoardState) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.add_new_state(
			p_game_id => $1,
			p_state => $2)`,
		bs.GameID, bs).
		Scan(&DBerror)
	if err != nil {
		return err
	}
	if DBerror.Status == pgtype.Present {
		switch DBerror.Int {
		case 1:
			return errors.New("[ERROR] Не удалось найти игру")
		default:
			return errors.New("[ERROR] Непредвиденная ошибка")
		}
	}
	return nil
}

func GetBoardStateFromDB(db *pgx.Conn, gameId int) (BoardState, error) {
	var bs BoardState
	var JSONBData string
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.get_board_state(
			p_game_id => $1)`,
		gameId).
		Scan(&JSONBData)
	if err != nil {
		return bs, err
	}
	err = json.Unmarshal([]byte(JSONBData), &bs)
	if err != nil {
		return bs, err
	}
	return bs, nil
}

func CompleteGameInDB(db *pgx.Conn, gameId int, userId int) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.complete_game(
			p_game_id => $1,
			p_user_id => $2)`,
		gameId, userId).
		Scan(&DBerror)
	if err != nil {
		return err
	}
	if DBerror.Status == pgtype.Present {
		switch DBerror.Int {
		case 1:
			return errors.New("[ERROR] Не удалось найти игру")
		default:
			return errors.New("[ERROR] Непредвиденная ошибка")
		}
	}
	return nil
}

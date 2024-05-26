package lobbies

import (
	"context"
	"errors"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5"
)

func CreateLobbyInDB(db *pgx.Conn, userId int) error {
	var lobbyId int
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.create_lobby(
			p_user_id => $1)`,
		userId).
		Scan(&lobbyId)
	if err != nil {
		return err
	}
	return nil
}

func GetLobbyInfoFromDB(db *pgx.Conn, lobbies *[]Lobby) error {
	rows, err := db.Query(context.Background(),
		`SELECT * FROM dbo.get_lobby_info()`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var lobby Lobby
		var player1, player2, player3, player4 Player

		err = rows.Scan(
			&lobby.LobbyId,
			&lobby.CreatorId,
			&lobby.Status,
			&lobby.CreationDate,
			&player1.UserId, &player1.Username, &player1.JoinedAt,
			&player2.UserId, &player2.Username, &player2.JoinedAt,
			&player3.UserId, &player3.Username, &player3.JoinedAt,
			&player4.UserId, &player4.Username, &player4.JoinedAt,
		)
		if err != nil {
			return err
		}

		lobby.Players = make([]Player, 0, 4)
		if player1.UserId != 0 {
			lobby.Players = append(lobby.Players, player1)
		}
		if player2.UserId != 0 {
			lobby.Players = append(lobby.Players, player2)
		}
		if player3.UserId != 0 {
			lobby.Players = append(lobby.Players, player3)
		}
		if player4.UserId != 0 {
			lobby.Players = append(lobby.Players, player4)
		}

		*lobbies = append(*lobbies, lobby)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}

func JoinLobbyInDB(db *pgx.Conn, lobbyId int, userId int) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.add_to_lobby(
			p_lobby_id => $1,
			p_user_id => $2)`,
		lobbyId, userId).
		Scan(&DBerror)
	if err != nil {
		return err
	}

	if DBerror.Status == pgtype.Present {
		switch DBerror.Int {
		case 1:
			return errors.New("[ERROR] Не удалось добавить в лобби")
		default:
			return errors.New("[ERROR] Непредвиденная ошибка")
		}
	}
	return nil
}

func LeaveLobbyInDB(db *pgx.Conn, lobbyId int, userId int) error {
	var DBerror pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.leave_lobby(
			p_lobby_id => $1,
			p_user_id => $2)`,
		lobbyId, userId).
		Scan(&DBerror)
	if err != nil {
		return err
	}

	if DBerror.Status == pgtype.Present {
		switch DBerror.Int {
		case 1:
			return errors.New("[ERROR] Не удалось выйти из лобби")
		default:
			return errors.New("[ERROR] Непредвиденная ошибка")
		}
	}
	return nil
}

func StartGameInDB(db *pgx.Conn, lobbyId int, creatorId int) (int, error) {
	var gameId pgtype.Int4
	err := db.QueryRow(context.Background(),
		`SELECT * FROM dbo.start_game(
			p_lobby_id => $1,
			p_creator_id => $2)`,
		lobbyId, creatorId).
		Scan(&gameId)
	if err != nil {
		return 0, err
	}
	if gameId.Status == pgtype.Null {
		return 0, errors.New("ошибка создания игры")
	}
	return int(gameId.Int), nil
}

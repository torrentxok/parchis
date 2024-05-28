package lobbies

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5"
	database "github.com/torrentxok/parchis/pkg/db"
	"github.com/torrentxok/parchis/pkg/game"
)

func CreateLobby(userId int) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = CreateLobbyInDB(db, userId)
	if err != nil {
		log.Print("[ERROR] Ошибка создания лобби: " + err.Error())
		return err
	}

	return nil
}

func GetLobbies() ([]Lobby, error) {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return nil, err
	}
	defer db.Close(context.Background())
	var lobbies []Lobby
	err = GetLobbyInfoFromDB(db, &lobbies)
	if err != nil {
		log.Print("[ERROR] Ошибка получения лобби: " + err.Error())
		return nil, err
	}
	return lobbies, nil
}

func JoinLobby(lobbyId int, userId int) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = JoinLobbyInDB(db, lobbyId, userId)
	if err != nil {
		log.Print("[ERROR] Ошибка добавления в лобби: " + err.Error())
		return err
	}

	return nil
}

func LeaveLobby(lobbyId int, userId int) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = LeaveLobbyInDB(db, lobbyId, userId)
	if err != nil {
		log.Print("[ERROR] Ошибка выхода из лобби: " + err.Error())
		return err
	}

	return nil
}

func StartGame(lobbyId int, creatorId int) (int, []int, error) {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return 0, nil, err
	}
	defer db.Close(context.Background())

	gameId, players, err := StartGameInDB(db, lobbyId, creatorId)
	if err != nil {
		log.Print("[ERROR] Ошибка создания игры: " + err.Error())
		return 0, nil, err
	}
	err = InitGame(db, gameId, players)
	if err != nil {
		log.Print("[ERROR] Ошибка инициализации игры: " + err.Error())
		return 0, nil, err
	}
	return gameId, players, nil
}

func InitGame(db *pgx.Conn, gameId int, playersIds []int) error {
	players := make(map[int]struct {
		UserId int
		Color  string
		Tokens map[int]struct {
			Position int
			CanMove  bool
		}
	})

	colors := []string{"yellow", "green", "red", "blue"}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for _, userId := range playersIds {
		tokens := make(map[int]struct {
			Position int
			CanMove  bool
		})

		for j := 0; j < 4; j++ {
			tokens[j] = struct {
				Position int
				CanMove  bool
			}{Position: -1, CanMove: false}
		}

		index := r.Intn(len(colors))
		selectedColor := colors[index]
		players[userId] = struct {
			UserId int
			Color  string
			Tokens map[int]struct {
				Position int
				CanMove  bool
			}
		}{
			UserId: userId,
			Color:  selectedColor,
			Tokens: tokens,
		}
		colors = append(colors[:index], colors[index+1:]...)
	}

	board := make(map[int]struct{ BlockedBy string })
	for i := 0; i <= 75; i++ {
		board[i] = struct{ BlockedBy string }{BlockedBy: ""}
	}
	bs := game.BoardState{
		GameID: gameId,
		Players: map[int]struct {
			UserId int    "json:\"player_id\""
			Color  string "json:\"color\""
			Tokens map[int]struct {
				Position int  "json:\"position\""
				CanMove  bool "json:\"can_move\""
			} "json:\"tokens\""
		}(players),
		CurrentTurn:  playersIds[r.Intn(len(playersIds))],
		DiceValue:    0,
		StrokeStatus: "roll_dice",
		Winner:       -1,
		Board: map[int]struct {
			BlockedBy string "json:\"blocked_by\""
		}(board),
	}
	err := game.AddNewStateToDB(db, bs)
	if err != nil {
		log.Print("[ERROR] Ошибка инициализации" + err.Error())
		return err
	}
	return nil
}

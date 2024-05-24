package lobbies

import (
	"context"
	"log"

	database "github.com/torrentxok/parchis/pkg/db"
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

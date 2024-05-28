package game

import (
	"context"
	"errors"
	"log"
	"sort"

	"github.com/thoas/go-funk"
	database "github.com/torrentxok/parchis/pkg/db"
)

func IsPlayerInGame(gameId int, userId int) bool {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return false
	}
	defer db.Close(context.Background())
	playerInGame, err := IsPlayerInGameDB(db, gameId, userId)
	if err != nil {
		log.Print("[ERROR] Ошибка проверки пользователя: " + err.Error())
		return false
	}
	return playerInGame
}

func IsExistsGame(gameId int) bool {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return false
	}
	defer db.Close(context.Background())
	isExistsGame, err := ExistsGame(db, gameId)
	if err != nil {
		log.Print("[ERROR] Ошибка поиска игры: " + err.Error())
		return false
	}
	return isExistsGame
}

func AddPlayerToGame(player *Player) error {
	GameMutex.Lock()
	game, exists := games[player.GameId]
	GameMutex.Unlock()
	if !exists {
		if IsExistsGame(player.GameId) {
			GameMutex.Lock()
			games[player.GameId] = &Game{GameId: player.GameId}
			GameMutex.Unlock()
		} else {
			log.Print("[ERROR] Игры не существует")
			return errors.New("игры не существует")
		}
	}

	game.Mutex.Lock()
	defer game.Mutex.Unlock()
	game.Players[player.UserId] = player
	return nil
}

func GetBoardState(gameId int) (BoardState, error) {
	var bs BoardState
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return bs, err
	}
	defer db.Close(context.Background())
	bs, err = GetBoardStateFromDB(db, gameId)
	if err != nil {
		log.Print("[ERROR] Ошибка получения данных: " + err.Error())
		return bs, err
	}
	return bs, nil
}

func CopyBoardState(original BoardState) BoardState {
	copy := original // Копируем все простые поля

	// Создаем новые карты для Players и Board
	copy.Players = make(map[int]struct {
		UserId int    "json:\"player_id\""
		Color  string "json:\"color\""
		Tokens map[int]struct {
			Position int  "json:\"position\""
			CanMove  bool "json:\"can_move\""
		} "json:\"tokens\""
	})
	copy.Board = make(map[int]struct {
		BlockedBy string "json:\"blocked_by\""
	})

	// Копируем данные из оригинальных карт в новые
	for k, v := range original.Players {
		tokensCopy := make(map[int]struct {
			Position int
			CanMove  bool
		})
		for tokenKey, tokenValue := range v.Tokens {
			tokensCopy[tokenKey] = struct {
				Position int
				CanMove  bool
			}(tokenValue)
		}
		v.Tokens = map[int]struct {
			Position int  "json:\"position\""
			CanMove  bool "json:\"can_move\""
		}(tokensCopy)
		copy.Players[k] = v
	}
	for k, v := range original.Board {
		copy.Board[k] = v
	}

	return copy
}

func GetPlayerColorsInOrder(bs BoardState) []string {
	// Создаем словарь для сопоставления цветов с их порядковыми номерами
	colorOrder := map[string]int{
		"yellow": 1,
		"green":  2,
		"red":    3,
		"blue":   4,
	}

	// Получаем цвета игроков
	playerColors := make([]string, 0, len(bs.Players))
	for _, player := range bs.Players {
		playerColors = append(playerColors, player.Color)
	}

	// Сортируем цвета игроков в соответствии с порядком colorOrder
	sort.Slice(playerColors, func(i, j int) bool {
		return colorOrder[playerColors[i]] < colorOrder[playerColors[j]]
	})

	return playerColors
}

func AddNewState(bs BoardState) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = AddNewStateToDB(db, bs)
	if err != nil {
		log.Print("[ERROR] Ошибка добавления: " + err.Error())
		return err
	}
	return nil
}

func RollDice(user *Player, bs BoardState) (BoardState, error) {
	dice := RandDice.Intn(6) + 1
	newbs := CopyBoardState(bs)
	newbs.DiceValue = dice
	newbs.StrokeStatus = "move"

	playerQueue := GetPlayerColorsInOrder(bs)
	playerColor := bs.Players[user.UserId].Color
	playerRoad := CellRoad[playerColor]
	for i, token := range bs.Players[user.UserId].Tokens {
		canMoveFlag := true
		if token.Position == -1 {
			if dice == 6 {
				canMoveFlag = true
			} else {
				canMoveFlag = false
			}
		} else {
			currentIndx := funk.IndexOf(playerRoad, token.Position)
			finishIndx := len(playerRoad) - 1
			if currentIndx+dice > finishIndx {
				canMoveFlag = false
			} else {
				for j := currentIndx; j <= currentIndx+dice; j++ {
					if bs.Board[playerRoad[j]].BlockedBy != "" &&
						bs.Board[playerRoad[j]].BlockedBy != playerColor {
						canMoveFlag = false
						break
					}
				}
				if canMoveFlag {
					for _, color := range playerQueue {
						if color == playerColor {
							continue
						}
						for v, otherToken := range bs.Players[user.UserId].Tokens {
							if v == i {
								continue
							}
							if playerRoad[currentIndx+dice] == StartCells[color] &&
								playerRoad[currentIndx+dice] == otherToken.Position {
								canMoveFlag = false
							} else {
								canMoveFlag = true
							}
						}
					}
				}
			}
		}
		newPlayer := newbs.Players[user.UserId]
		newToken := newPlayer.Tokens[i]
		newToken.CanMove = canMoveFlag
		newPlayer.Tokens[i] = newToken
		newbs.Players[user.UserId] = newPlayer
	}
	return newbs, nil
}

func IsCellProtected(bs BoardState, newPos int, userId int) bool {
	counter := 0
	for _, token := range bs.Players[userId].Tokens {
		if token.Position == newPos {
			counter++
		}
	}
	return counter >= 2
}

func CheckWinner(bs BoardState) int {
	for id, user := range bs.Players {
		counter := 0
		finishPos := FinishCell[user.Color]
		for _, token := range user.Tokens {
			if token.Position == finishPos {
				counter++
			}
		}
		if counter == 4 {
			return id
		}
	}
	return -1
}

func MoveToken(user *Player, bs BoardState, TokenId int) (BoardState, error) {
	newbs := CopyBoardState(bs)
	newbs.StrokeStatus = "roll_dice"
	oldPosition := bs.Players[user.UserId].Tokens[TokenId].Position
	var newPosition int
	playerQueue := GetPlayerColorsInOrder(bs)
	playerColor := bs.Players[user.UserId].Color
	playerRoad := CellRoad[playerColor]
	// добавить логику добавления защиты и освобождения от защиты
	// добавить логику когда нет ходов
	if TokenId == -1 {
		for _, token := range bs.Players[user.UserId].Tokens {
			if token.CanMove {
				return newbs, errors.New("игрок может сделать ход")
			}
		}
	} else {
		if oldPosition == -1 {
			newPlayer := newbs.Players[user.UserId]
			newToken := newPlayer.Tokens[TokenId]
			newPosition = StartCells[newPlayer.Color]
			newToken.Position = newPosition
			newPlayer.Tokens[TokenId] = newToken
			newbs.Players[user.UserId] = newPlayer
		} else {
			newPlayer := newbs.Players[user.UserId]
			newToken := newPlayer.Tokens[TokenId]
			currentIndx := funk.IndexOf(playerRoad, newToken.Position)
			newPosition = playerRoad[currentIndx+newbs.DiceValue]
			newToken.Position = newPosition
			newPlayer.Tokens[TokenId] = newToken
			newbs.Players[user.UserId] = newPlayer
		}
		for id, player := range bs.Players {
			if id == user.UserId {
				continue
			}
			for k, token := range player.Tokens {
				if token.Position == newPosition {
					newOtherPlayer := newbs.Players[id]
					newOtherToken := newOtherPlayer.Tokens[k]
					newOtherToken.Position = -1
					newOtherPlayer.Tokens[k] = newOtherToken
					newbs.Players[id] = newOtherPlayer
				}
			}
		}
		newBoard1 := newbs.Board[newPosition]
		if IsCellProtected(newbs, newPosition, user.UserId) {
			newBoard1.BlockedBy = newbs.Players[user.UserId].Color
		} else {
			newBoard1.BlockedBy = ""
		}
		newbs.Board[newPosition] = newBoard1

		newBoard2 := newbs.Board[newPosition]
		if IsCellProtected(newbs, oldPosition, user.UserId) {
			newBoard2.BlockedBy = newbs.Players[user.UserId].Color
		} else {
			newBoard2.BlockedBy = ""
		}
		newbs.Board[newPosition] = newBoard2
	}

	currentPlayerIndex := funk.IndexOf(playerQueue, newbs.Players[user.UserId].Color)
	var newColor string
	if currentPlayerIndex == len(playerQueue)-1 {
		newColor = playerQueue[0]
	} else {
		newColor = playerQueue[currentPlayerIndex+1]
	}
	for id, player := range bs.Players {
		if player.Color == newColor {
			newbs.CurrentTurn = id
		}
	}
	return newbs, nil
}

func CompleteGame(winner int, bs BoardState) error {
	db, err := database.ConnectToDB()
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к базе данных: " + err.Error())
		return err
	}
	defer db.Close(context.Background())
	err = CompleteGameInDB(db, bs.GameID, winner)
	if err != nil {
		log.Print("[ERROR] Ошибка завершения игры: " + err.Error())
		return err
	}
	return nil
}

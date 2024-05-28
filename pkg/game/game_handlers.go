package game

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/torrentxok/parchis/pkg/api"
	"github.com/torrentxok/parchis/pkg/auth"
)

func GameHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	gameId, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Print("[ERROR] Invalid profile ID: " + err.Error())
		api.SendErrorResponse(w, "Invalid profile ID", http.StatusBadRequest)
		return
	}
	claims, ok := r.Context().Value(auth.ClaimsKey).(jwt.MapClaims)
	if !ok {
		log.Print("[ERROR] No claims found")
		api.SendErrorResponse(w, "No claims found", http.StatusInternalServerError)
		return
	}
	userIdFloat64, ok := claims["UserId"].(float64)
	if !ok {
		log.Print("[ERROR] Invalid user ID in claims")
		api.SendErrorResponse(w, "Invalid user ID in claims", http.StatusInternalServerError)
		return
	}
	userId := int(userIdFloat64)

	if !IsPlayerInGame(gameId, userId) {
		log.Print("[ERROR] Ошибка подключения: " + err.Error())
		api.SendErrorResponse(w, "Ошибка подключения", http.StatusForbidden)
		return
	}

	conn, err := GameUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к websocket")
		api.SendErrorResponse(w, "Ошибка подключения к websocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	player := &Player{
		conn:   conn,
		UserId: userId,
		GameId: gameId,
	}
	err = AddPlayerToGame(player)
	if err != nil {
		log.Print("[ERROR] Ошибка присоединения к игре")
		api.SendErrorResponse(w, "Ошибка присоединения к игре", http.StatusBadRequest)
		return
	}

	for {
		var gameReq GameMessageReq
		err = conn.ReadJSON(&gameReq)
		if err != nil {
			log.Print("[ERROR] Ошибка чтения")
			GameMutex.Lock()
			delete(games, player.GameId)
			GameMutex.Unlock()
			conn.Close()
			break
		}

		var gameResp GameMessageResp
		switch gameReq.Type {
		case "roll_dice":
			gameResp, err = RollDiceHandler(player, games[player.GameId])
			if err != nil {
				log.Print("[ERROR] Ошибка броска" + err.Error())
			}
		case "move":
			gameResp, err = MoveTokenHandler(player, games[player.GameId], gameReq.Data)
			if err != nil {
				log.Print("[ERROR] Ошибка перемещения" + err.Error())
			}
		case "get_state":
			gameResp, err = GetStateHandler(player.GameId)
			if err != nil {
				log.Print("[ERROR] Ошибка получения игры" + err.Error())
			}
		default:
		}

		if err != nil {
			err := player.conn.WriteJSON(gameResp)
			if err != nil {
				log.Print("[ERROR] Ошибка записи")
				delete(games[player.GameId].Players, player.UserId)
				player.conn.Close()
			}
		} else {
			BroadcastToPlayers(gameResp, games[player.GameId])
		}
	}
}

func BroadcastToPlayers(msg GameMessageResp, game *Game) {
	game.Mutex.Lock()
	defer game.Mutex.Unlock()

	for id, user := range game.Players {
		err := user.conn.WriteJSON(msg)
		if err != nil {
			log.Print("[ERROR] Ошибка записи")
			delete(game.Players, id)
			user.conn.Close()
		}
	}
}

func GetStateHandler(gameId int) (GameMessageResp, error) {
	var resp GameMessageResp
	bs, err := GetBoardState(gameId)
	if err != nil {
		log.Print("[ERROR] Ошибка получения игры")
		resp.Code = http.StatusBadRequest
		resp.Message = "Ошибка получения игры"
		resp.Status = "error"
		resp.Data = nil
		return resp, err
	}
	resp.Code = http.StatusOK
	resp.Message = ""
	resp.Status = "success"
	resp.Type = "board_state"
	resp.Data = bs
	return resp, nil
}

func RollDiceHandler(user *Player, game *Game) (GameMessageResp, error) {
	var resp GameMessageResp
	bs, err := GetBoardState(game.GameId)
	if err != nil {
		log.Print("[ERROR] Ошибка получения игры")
		resp.Code = http.StatusBadRequest
		resp.Message = "Ошибка получения игры"
		resp.Status = "error"
		resp.Data = nil
		return resp, err
	}
	var newBS BoardState
	if bs.CurrentTurn == user.UserId && bs.StrokeStatus == "roll_dice" {
		newBS, err = RollDice(user, bs)
		if err != nil {
			log.Print("[ERROR] Ошибка броска кубика")
			resp.Code = http.StatusBadRequest
			resp.Message = "Ошибка броска кубика"
			resp.Status = "error"
			resp.Data = nil
			return resp, err
		}
		err = AddNewState(newBS)
		if err != nil {
			log.Print("[ERROR] Ошибка добавления состояния")
			resp.Code = http.StatusBadRequest
			resp.Message = "Ошибка добавления состояния"
			resp.Status = "error"
			resp.Data = nil
			return resp, err
		}
	}
	resp.Code = http.StatusOK
	resp.Message = ""
	resp.Status = "success"
	resp.Type = "board_state"
	resp.Data = newBS
	return resp, nil
}

func MoveTokenHandler(user *Player, game *Game, data json.RawMessage) (GameMessageResp, error) {
	var resp GameMessageResp
	var req MoveTokenReq
	err := json.Unmarshal([]byte(data), &req)
	if err != nil {
		log.Print("[ERROR] Ошибка при разборе данных: " + err.Error())
		resp.Code = http.StatusBadRequest
		resp.Message = "Ошибка при разборе данных"
		resp.Status = "error"
		resp.Data = nil
		return resp, errors.New("ошибка при разборе данных")
	}
	bs, err := GetBoardState(game.GameId)
	if err != nil {
		log.Print("[ERROR] Ошибка получения игры")
		resp.Code = http.StatusBadRequest
		resp.Message = "Ошибка получения игры"
		resp.Status = "error"
		resp.Data = nil
		return resp, err
	}
	var newBS BoardState
	// логика
	if bs.CurrentTurn == user.UserId && bs.StrokeStatus == "move" && bs.Players[user.UserId].Tokens[req.TokenId].CanMove {
		newBS, err = MoveToken(user, bs, req.TokenId)
		if err != nil {
			log.Print("[ERROR] Ошибка хода")
			resp.Code = http.StatusBadRequest
			resp.Message = "Ошибка хода"
			resp.Status = "error"
			resp.Data = nil
			return resp, err
		}
		err = AddNewState(newBS)
		if err != nil {
			log.Print("[ERROR] Ошибка добавления состояния")
			resp.Code = http.StatusBadRequest
			resp.Message = "Ошибка добавления состояния"
			resp.Status = "error"
			resp.Data = nil
			return resp, err
		}
		winner := CheckWinner(newBS)
		if winner != -1 {
			err = CompleteGame(winner, newBS)
			if err != nil {
				log.Print("[ERROR] Ошибка завершения игры")
				resp.Code = http.StatusBadRequest
				resp.Message = "Ошибка завершения игры"
				resp.Status = "error"
				resp.Data = nil
				return resp, err
			}
			resp.Code = http.StatusOK
			resp.Message = ""
			resp.Status = "success"
			resp.Type = "winner"
			resp.Data = winner
			return resp, nil
		}
	}

	resp.Code = http.StatusOK
	resp.Message = ""
	resp.Status = "success"
	resp.Type = "board_state"
	resp.Data = newBS
	return resp, nil
}

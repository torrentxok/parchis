package lobbies

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/torrentxok/parchis/pkg/api"
)

func LobbiesHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("[ERROR] Ошибка подключения к websocket")
		api.SendErrorResponse(w, "Ошибка подключения к websocket", http.StatusBadRequest)
		return
	}

	client := &Client{conn: conn}
	mutex.Lock()
	clients[client] = true
	mutex.Unlock()

	for {
		var requestMsg MessageReq
		err = conn.ReadJSON(&requestMsg)
		if err != nil {
			log.Print("[ERROR] Ошибка чтения")
			mutex.Lock()
			delete(clients, client)
			mutex.Unlock()
			conn.Close()
			break
		}

		var responseMsg MessageResp

		switch requestMsg.Type {
		case "create_lobby":
			responseMsg, err = CreateLobbyHandler(requestMsg.Data)
			if err != nil {
				log.Print("[ERROR] Ошибка создания лобби" + err.Error())
			}
		case "get_lobbies":
			responseMsg, err = GetLobbiesHandler()
			if err != nil {
				log.Print("[ERROR] Ошибка получения лобби" + err.Error())
			}
		case "join_lobby":
			responseMsg, err = JoinLobbyHandler(requestMsg.Data)
			if err != nil {
				log.Print("[ERROR] Ошибка добавления в лобби" + err.Error())
			}
		case "leave_lobby":
			responseMsg, err = LeaveLobbyHandler(requestMsg.Data)
			if err != nil {
				log.Print("[ERROR] Ошибка выхода из лобби" + err.Error())
			}
		default:
			log.Print("[ERROR] Ошибка декодирования")
			err = errors.New("ошибка декодирования")
			responseMsg.Code = http.StatusBadRequest
			responseMsg.Message = "Ошибка декодирования"
			responseMsg.Status = "error"
			responseMsg.Data = nil
		}

		if err != nil {
			err := client.conn.WriteJSON(responseMsg)
			if err != nil {
				log.Print("[ERROR] Ошибка записи")
				delete(clients, client)
				client.conn.Close()
			}
		} else {
			Broadcast(responseMsg)
		}

		// обработка сообщения

		// err = conn.WriteJSON(msg)
		// if err != nil {
		// 	log.Print("[ERROR] Ошибка записи")
		// 	return
		// }
	}
}

func Broadcast(msg MessageResp) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range clients {
		err := client.conn.WriteJSON(msg)
		if err != nil {
			log.Print("[ERROR] Ошибка записи")
			delete(clients, client)
			client.conn.Close()
		}
	}
}

func CreateLobbyHandler(data json.RawMessage) (MessageResp, error) {
	var crReq CreateLobbyReq
	var crResp MessageResp
	err := json.Unmarshal([]byte(data), &crReq)
	if err != nil {
		log.Print("[ERROR] Ошибка при разборе данных: " + err.Error())
		crResp.Code = http.StatusBadRequest
		crResp.Message = "Ошибка при разборе данных"
		crResp.Status = "error"
		crResp.Data = nil
		return crResp, errors.New("ошибка при разборе данных")
	}
	err = CreateLobby(crReq.UserId)
	if err != nil {
		log.Print("[ERROR] Ошибка создания лобби: " + err.Error())
		crResp.Code = http.StatusBadRequest
		crResp.Message = "Ошибка создания лобби"
		crResp.Status = "error"
		crResp.Data = nil
		return crResp, errors.New("ошибка создания лобби")
	}

	lobbies, err := GetLobbies()
	if err != nil {
		log.Print("[ERROR] Ошибка получения лобби")
		crResp.Code = http.StatusBadRequest
		crResp.Message = "Ошибка получения лобби"
		crResp.Status = "error"
		crResp.Data = nil
		return crResp, err
	}

	crResp.Code = http.StatusOK
	crResp.Message = ""
	crResp.Status = "success"
	crResp.Data = lobbies

	return crResp, nil
}

func GetLobbiesHandler() (MessageResp, error) {
	var glResp MessageResp
	lobbies, err := GetLobbies()
	if err != nil {
		log.Print("[ERROR] Ошибка получения лобби")
		glResp.Code = http.StatusBadRequest
		glResp.Message = "Ошибка получения лобби"
		glResp.Status = "error"
		glResp.Data = nil
		return glResp, err
	}

	glResp.Code = http.StatusOK
	glResp.Message = ""
	glResp.Status = "success"
	glResp.Data = lobbies

	return glResp, nil
}

func JoinLobbyHandler(data json.RawMessage) (MessageResp, error) {
	var jlReq JoinLobbyReq
	var jlResp MessageResp
	err := json.Unmarshal([]byte(data), &jlReq)
	if err != nil {
		log.Print("[ERROR] Ошибка при разборе данных: " + err.Error())
		jlResp.Code = http.StatusBadRequest
		jlResp.Message = "Ошибка при разборе данных"
		jlResp.Status = "error"
		jlResp.Data = nil
		return jlResp, errors.New("ошибка при разборе данных")
	}

	err = JoinLobby(jlReq.LobbyId, jlReq.UserId)
	if err != nil {
		log.Print("[ERROR] Ошибка присоединения к лобби: " + err.Error())
		jlResp.Code = http.StatusBadRequest
		jlResp.Message = "Ошибка присоединения к лобби"
		jlResp.Status = "error"
		jlResp.Data = nil
		return jlResp, errors.New("ошибка присоединения к лобби")
	}

	lobbies, err := GetLobbies()
	if err != nil {
		log.Print("[ERROR] Ошибка получения лобби")
		jlResp.Code = http.StatusBadRequest
		jlResp.Message = "Ошибка получения лобби"
		jlResp.Status = "error"
		jlResp.Data = nil
		return jlResp, err
	}

	jlResp.Code = http.StatusOK
	jlResp.Message = ""
	jlResp.Status = "success"
	jlResp.Data = lobbies
	return jlResp, nil
}

func LeaveLobbyHandler(data json.RawMessage) (MessageResp, error) {
	var llReq LeaveLobbyReq
	var llResp MessageResp
	err := json.Unmarshal([]byte(data), &llReq)
	if err != nil {
		log.Print("[ERROR] Ошибка при разборе данных: " + err.Error())
		llResp.Code = http.StatusBadRequest
		llResp.Message = "Ошибка при разборе данных"
		llResp.Status = "error"
		llResp.Data = nil
		return llResp, errors.New("ошибка при разборе данных")
	}

	err = LeaveLobby(llReq.LobbyId, llReq.UserId)
	if err != nil {
		log.Print("[ERROR] Ошибка выхода из лобби: " + err.Error())
		llResp.Code = http.StatusBadRequest
		llResp.Message = "Ошибка выхода из лобби"
		llResp.Status = "error"
		llResp.Data = nil
		return llResp, errors.New("ошибка выхода из лобби")
	}

	lobbies, err := GetLobbies()
	if err != nil {
		log.Print("[ERROR] Ошибка получения лобби")
		llResp.Code = http.StatusBadRequest
		llResp.Message = "Ошибка получения лобби"
		llResp.Status = "error"
		llResp.Data = nil
		return llResp, err
	}

	llResp.Code = http.StatusOK
	llResp.Message = ""
	llResp.Status = "success"
	llResp.Data = lobbies
	return llResp, nil
}

package lobbies

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  8 * 1024, // 8 килобайта
	WriteBufferSize: 8 * 1024, // 8 килобайта
}

var clients = make(map[*Client]bool)
var mutex = &sync.Mutex{}

type Client struct {
	conn *websocket.Conn
}

type MessageReq struct {
	Type string          `json:"type"`
	Data json.RawMessage `jaon:"data"`
}

type MessageResp struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

type CreateLobbyReq struct {
	UserId int `json:"user_id"`
}

type Player struct {
	UserId   int       `json:"user_id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

type Lobby struct {
	LobbyId      int       `json:"lobby_id"`
	CreatorId    int       `json:"creator_id"`
	Status       string    `json:"status"`
	CreationDate time.Time `json:"creation_date"`
	Players      []Player  `json:"players"`
}

type JoinLobbyReq struct {
	LobbyId int `json:"lobby_id"`
	UserId  int `json:"user_id"`
}

type LeaveLobbyReq struct {
	LobbyId int `json:"lobby_id"`
	UserId  int `json:"user_id"`
}

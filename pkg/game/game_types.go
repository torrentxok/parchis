package game

import (
	"encoding/json"
	"math/rand"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	GameMutex = &sync.Mutex{}
	RandDice  = rand.New(rand.NewSource(time.Now().UnixNano()))
	//players      = make(map[*Player]bool)
	games        = make(map[int]*Game)
	GameUpgrader = websocket.Upgrader{
		ReadBufferSize:  8 * 1024, // 8 килобайта
		WriteBufferSize: 8 * 1024, // 8 килобайта
	}
	DefaultQueue = []string{"yellow", "green", "red", "blue"}
	StartCells   = map[string]int{
		"yellow": 0,
		"green":  13,
		"red":    26,
		"blue":   39,
	}
	FinishCell = map[string]int{
		"yellow": 57,
		"green":  63,
		"red":    69,
		"blue":   75,
	}
	HomeCells = map[string][]int{
		"yellow": {52, 53, 54, 55, 56},
		"green":  {58, 59, 60, 61, 62},
		"red":    {64, 65, 66, 67, 68},
		"blue":   {70, 71, 72, 73, 74},
	}
	CellRoad = map[string][]int{
		"yellow": {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19,
			20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39,
			40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 52, 53, 54, 55, 56, 57},
		"green": {13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32,
			33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 0,
			1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 58, 59, 60, 61, 62, 63},
		"red": {26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45,
			46, 47, 48, 49, 50, 51, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
			14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 64, 65, 66, 67, 68, 69},
		"blue": {39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 0, 1, 2, 3, 4, 5, 6,
			7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
			27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 70, 71, 72, 73, 74, 75},
	}
)

type Player struct {
	conn   *websocket.Conn
	UserId int
	GameId int
	Color  string
}

type Game struct {
	GameId  int
	Players map[int]*Player
	Mutex   sync.Mutex
}

type BoardState struct {
	GameID  int `json:"game_id"`
	Players map[int]struct {
		UserId int    `json:"player_id"`
		Color  string `json:"color"`
		Tokens map[int]struct {
			Position int  `json:"position"`
			CanMove  bool `json:"can_move"`
		} `json:"tokens"`
	} `json:"players"`
	CurrentTurn  int    `json:"current_turn"`
	DiceValue    int    `json:"dice_value"`
	StrokeStatus string `json:"stroke_status"`
	Winner       int    `json:"winner"`
	Board        map[int]struct {
		BlockedBy string `json:"blocked_by"`
	} `json:"board"`
}

type GameMessageReq struct {
	Type string          `json:"type"`
	Data json.RawMessage `jaon:"data"`
}

type GameMessageResp struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
}

type MoveTokenReq struct {
	TokenId int `json:"token_id"`
}

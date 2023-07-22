package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"

	"github.com/anthdm/hollywood/actor"
	"github.com/gorilla/websocket"
	"github.com/morka17/fireship/models"
)

type PlayerState struct {
	Position models.Position
	Health int 
	Velocity int 
}


type PlayerSession struct {
	sessionID int
	clientID  int
	username  string
	inLobby   bool
	conn      *websocket.Conn
}

func newPlayerSession(sid int, conn *websocket.Conn) actor.Producer {
	return func() actor.Receiver {
		return &PlayerSession{
			sessionID: sid,
			conn:      conn,
		}
	}
}

func (s *PlayerSession) Receive(c *actor.Context) {
	switch c.Message().(type) {
	case actor.Started:
		s.readLoop()
		// statePID = c.SpawnChild(newPlayerState, "playerState")
	}
}

func (s *PlayerSession) readLoop() {
	var msg models.WSMessage
	for {
		if err := s.conn.ReadJSON(&msg); err != nil {
			fmt.Println("read error", err)
			return
		}
		go s.handleMessage(msg)
	}
}

func (s *PlayerSession) handleMessage(msg models.WSMessage) {
	switch msg.Type {
	case "Login":
		var login models.Login
		if err := json.Unmarshal(msg.Data, &login); err != nil {
			panic(err)
		}
		fmt.Println(login)
		s.clientID = login.ClientID
		s.username = login.Username
	case "playerState":
		var ps models.PlayerState
		if err := json.Unmarshal(msg.Data, &ps); err != nil {
			panic(err)
		}
		fmt.Println(ps)
	}

}

type GameServer struct {
	ctx      *actor.Context
	sessions map[*actor.PID]struct{}
}

func NewGameServer() actor.Receiver {
	return &GameServer{
		sessions: make(map[*actor.PID]struct{}),
	}
}

func (s *GameServer) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		s.StartHTTP()
		s.ctx = c
		_ = msg
	}
}

func (s *GameServer) StartHTTP() {
	go func() {
		http.HandleFunc("/ws", s.HandleWS)
		http.ListenAndServe(":40000", nil)
	}()
}

// Handles the upgrade of the websocket
func (s *GameServer) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if err != nil {
		fmt.Println("ws upgrade err:", err)
		return
	}
	fmt.Print("New client trying connect")
	sid := rand.Intn(math.MaxInt)
	pid := s.ctx.SpawnChild(newPlayerSession(sid, conn), fmt.Sprintf("session_%d", sid))
	s.sessions[pid] = struct{}{}
	fmt.Printf("Client with sid %d and pid %s hust connected\n", sid, pid)
}

func main() {

	e := actor.NewEngine()
	e.Spawn(NewGameServer, "server")
	select{}
}

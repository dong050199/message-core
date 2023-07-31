package websocket

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// time to read the next client's pong message
	pongWait = 60 * time.Second
	// time period to send pings to client
	pingPeriod = (pongWait * 9) / 10
	// time allowed to write a message to client
	writeWait = 10 * time.Second
	// max message size allowed
	maxMessageSize = 512
	// I/O read buffer size
	readBufferSize = 1024
	// I/O write buffer size
	writeBufferSize = 1024
)

// Initialize server with empty subscription
var server = &Server{Subscriptions: make(Subscription)}

func GetServerConn() *Server {
	return server
}

// http to websocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  readBufferSize,
	WriteBufferSize: writeBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		// allow all origin
		return true
	},
}

func HandleWS(w http.ResponseWriter, r *http.Request) {
	// upgrades connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed upgrading connection"))
		return
	}
	defer conn.Close()

	// create new client id
	clientID := uuid.New().String()

	// greet the new client
	topic := r.URL.Query().Get("topic")
	if len(topic) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing parameters topic name"))
		return
	}

	server.Subscribe(conn, clientID, topic)

	// create channel to signal client health
	done := make(chan struct{})

	go writePump(conn, clientID, done)
	readPump(conn, clientID, done, topic)
}

// readPump process incoming messages and set the settings
func readPump(conn *websocket.Conn, clientID string, done chan<- struct{}, topic string) {
	// set limit, deadline to read & pong handler
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// message handling
	for {
		// read incoming message
		_, _, err := conn.ReadMessage()
		// if error occured
		if err != nil {
			// remove from the client
			server.RemoveClient(clientID)
			// set health status to unhealthy by closing channel
			close(done)
			// stop process
			break
		}
		// if no error, process incoming message
	}
}

// writePump sends ping to the client
func writePump(conn *websocket.Conn, clientID string, done <-chan struct{}) {
	// create ping ticker
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// send ping message
			err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait))
			if err != nil {
				// if error sending ping, remove this client from the server
				server.RemoveClient(clientID)
				// stop sending ping
				return
			}
		case <-done:
			// if process is done, stop sending ping
			return
		}
	}
}

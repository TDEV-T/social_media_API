package functional

import (
	"fmt"
	"log"
	"sync"
	"tdev/models"

	"github.com/gofiber/contrib/websocket"
)

type ChatServer struct {
	mu      sync.Mutex
	Clients map[*websocket.Conn]struct{}
}

func (cs *ChatServer) AddClient(c *websocket.Conn) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Clients[c] = struct{}{}
}

func (cs *ChatServer) RemoveClient(c *websocket.Conn) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.Clients, c)
}

func (cs *ChatServer) Broadcast(mt int, msg []byte) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	for conn := range cs.Clients {
		err := conn.WriteMessage(mt, msg)

		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func MessageSocket(cs *ChatServer) func(c *websocket.Conn) {

	return func(c *websocket.Conn) {

		userLocal := c.Locals("user").(*models.User)

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				break
			}

			defer c.Close()
			cs.AddClient(c)
			defer cs.RemoveClient(c)

			fmt.Printf("User : %s , Received message: %s\n", userLocal.Username, msg)

			cs.Broadcast(mt, msg)

		}
	}
}

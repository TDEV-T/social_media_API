package functional

import (
	"fmt"
	"log"
	"sync"
	"tdev/models"

	"github.com/gofiber/contrib/websocket"
)

type ChatServer struct {
	mu            sync.Mutex
	Clients       map[*websocket.Conn]struct{}
	Conversations map[string]map[*websocket.Conn]struct{} // New map to track conversations
}

func (cs *ChatServer) AddClient(c *websocket.Conn, conversationID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// สร้าง map ถ้ายังไม่มี
	if cs.Conversations == nil {
		cs.Conversations = make(map[string]map[*websocket.Conn]struct{})
	}

	// สร้าง map สำหรับ conversation ถ้ายังไม่มี
	if cs.Conversations[conversationID] == nil {
		cs.Conversations[conversationID] = make(map[*websocket.Conn]struct{})
	}

	// เพิ่ม client เข้าไปใน conversation
	cs.Conversations[conversationID][c] = struct{}{}
	cs.Clients[c] = struct{}{}
}

func (cs *ChatServer) RemoveClient(c *websocket.Conn, conversationID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	delete(cs.Conversations[conversationID], c)
	delete(cs.Clients, c)
}

func (cs *ChatServer) Broadcast(conversationID string, mt int, msg []byte) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for conn := range cs.Conversations[conversationID] {
		err := conn.WriteMessage(mt, msg)
		if err != nil {
			log.Println(err)
			continue
		}
	}
}

func MessageSocket(cs *ChatServer) func(c *websocket.Conn) {

	return func(c *websocket.Conn) {
		conId := c.Params("rid")

		userLocal := c.Locals("user").(*models.User)

		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				break
			}

			cs.AddClient(c, conId)
			defer cs.RemoveClient(c, conId)
			defer c.Close()

			fmt.Printf("User : %s , Received message: %s\n", userLocal.Username, msg)

			cs.Broadcast(conId, mt, msg)

		}
	}
}

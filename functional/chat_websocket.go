package functional

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"tdev/models"

	"github.com/gofiber/contrib/websocket"
	"gorm.io/gorm"
)

type ChatServer struct {
	mu            sync.Mutex
	Clients       map[*websocket.Conn]struct{}
	Conversations map[string]map[*websocket.Conn]struct{}
}

func (cs *ChatServer) AddClient(c *websocket.Conn, conversationID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.Conversations == nil {
		cs.Conversations = make(map[string]map[*websocket.Conn]struct{})
	}

	if cs.Conversations[conversationID] == nil {
		cs.Conversations[conversationID] = make(map[*websocket.Conn]struct{})
	}

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

func GetAllChatRoomWithUserID(db *gorm.DB) func(c *websocket.Conn) {
	return func(c *websocket.Conn) {
		userLocal := c.Locals("user").(*models.User)

		rooms, err := models.GetAllChatWithUserID(db, userLocal.ID)

		if err != nil {
			log.Println(err)
			c.WriteMessage(websocket.TextMessage, []byte("Error : "+err.Error()))
			c.Close()
			return
		}

		// if err := c.WriteMessage(websocket.TextMessage, []byte("Connect Success")); err != nil {
		// 	log.Println(err)
		// 	return
		// }

		for _, room := range rooms {
			roomJson, err := json.Marshal(room)

			if err != nil {
				log.Println(err)
				continue
			}

			if err := c.WriteMessage(websocket.TextMessage, roomJson); err != nil {
				log.Println(err)
				continue
			}
		}

	}
}

func JoinRoomChat(db *gorm.DB, cs *ChatServer) {

}

func MessageSocket(db *gorm.DB, cs *ChatServer) func(c *websocket.Conn) {

	return func(c *websocket.Conn) {
		receiverID, err := strconv.Atoi(c.Locals("receiverID").(string))

		if err != nil {
			log.Println(err)
			c.WriteMessage(websocket.TextMessage, []byte("Error : Invalid Receiver ID"))
			c.Close()
			return
		}

		userLocal := c.Locals("user").(*models.User)

		checkChatExists, err, rid := models.ChatRoomExists(db, userLocal.ID, uint(receiverID))

		if err != nil {
			log.Println(err)
			c.WriteMessage(websocket.TextMessage, []byte("Error :"+err.Error()))
			c.Close()
			return
		}

		if !checkChatExists {
			result, err := models.CreateChatRoom(db, userLocal.ID, uint(receiverID))

			if err != nil {
				log.Println(err)
				c.Close()
				return
			}

			rid = result.ID
		} else {
			messages, err := models.GetChatDetail(db, rid)

			if err != nil {
				log.Println(err)
				c.WriteMessage(websocket.TextMessage, []byte("Error :"+err.Error()))
				c.Close()
				return
			}

			for _, msg := range messages {
				c.WriteMessage(websocket.TextMessage, []byte(msg.Message))
			}
		}

		conId := strconv.Itoa(int(rid))

		cs.AddClient(c, conId)
		defer cs.RemoveClient(c, conId)
		defer c.Close()

		for {
			mt, msg, err := c.ReadMessage()

			if err != nil {
				log.Println("read error:", err)
				break
			}

			_, err = models.CreateMessage(db, userLocal.ID, string(msg), rid)

			if err != nil {
				log.Println("Create error:", err)
				c.WriteMessage(websocket.TextMessage, []byte("Error : "+err.Error()))
				c.Close()
				break
			}

			fmt.Printf("User : %s , Received message: %s\n", userLocal.Username, msg)

			cs.Broadcast(conId, mt, msg)

		}
	}
}

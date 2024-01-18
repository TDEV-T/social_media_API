package models

import (
	"fmt"

	"gorm.io/gorm"
)

type ChatRoom struct {
	gorm.Model
	Members []User `gorm:"many2many:chat_room_members;"`
}

type ChatMessage struct {
	gorm.Model
	SenderID uint
	RoomID   uint
	Message  string

	Sender User `gorm:"foreignkey:SenderID"`
}

func (cr *ChatRoom) TableName() string {
	return "chat_rooms"
}

func (cm *ChatMessage) TableName() string {
	return "chat_message"
}

func ChatRoomExists(db *gorm.DB, user1ID, user2ID uint) (bool, error, uint) {
	RoomResult := new(ChatRoom)

	err := db.Table("chat_rooms").Joins("JOIN chat_room_members on chat_room_members.chat_room_id = chat_rooms.id").Where("chat_room_members.user_id IN (?)", []uint{user1ID, user2ID}).Group("chat_rooms.id").Having("COUNT(DISTINCT chat_room_members.user_id) = ? ", 2).First(RoomResult).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, 0
		}
		return false, err, 0
	}

	return true, nil, RoomResult.ID
}

func GetChatDetail(db *gorm.DB, rid uint) ([]ChatMessage, error) {
	var messages []ChatMessage

	err := db.Where("room_id = ?", rid).Preload("Sender").Find(&messages).Error
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func GetAllChatWithUserID(db *gorm.DB, userID uint) ([]ChatRoom, error) {
	var rooms []ChatRoom

	err := db.Joins("JOIN chat_room_members on chat_room_members.chat_room_id = chat_rooms.id").Where("chat_room_members.user_id = ?", userID).Find(&rooms).Error

	if err != nil {
		return nil, err
	}

	return rooms, nil

}

func CreateChatRoom(db *gorm.DB, u1 uint, u2 uint) (ChatRoom, error) {
	var room ChatRoom

	var user1, user2 User

	if err := db.First(&user1, u1).Error; err != nil {
		return room, err
	}

	if err := db.First(&user2, u2).Error; err != nil {
		return room, err
	}

	if err := db.Create(&room).Error; err != nil {
		return room, err
	}

	fmt.Printf("ChatRoom : %d \n", room.ID)

	if err := db.Model(&room).Association("Members").Append([]User{user1, user2}).Error; err != nil {
		return room, nil
	}

	return room, nil
}

func CreateMessage(db *gorm.DB, userID uint, msg string, conId uint) (ChatMessage, error) {
	var chatmsg ChatMessage

	var user User
	var chatR ChatRoom

	if err := db.First(&user, userID).Error; err != nil {
		return chatmsg, err
	}

	if err := db.First(&chatR, conId).Error; err != nil {
		return chatmsg, err
	}

	chatmsg.SenderID = user.ID
	chatmsg.Message = msg
	chatmsg.RoomID = chatR.ID

	if err := db.Save(&chatmsg).Error; err != nil {
		return chatmsg, err
	}

	return chatmsg, nil

}

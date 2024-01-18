package models

import (
	"github.com/gofiber/fiber/v2"
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

	sender User `gorm:"foreignkey:SenderID"`
}

type RoomExistsResult struct {
	count  int
	roomID uint
}

func (cr *ChatRoom) TableName() string {
	return "chat_rooms"
}

func (cm *ChatMessage) TableName() string {
	return "chat_message"
}

func CreateTable(db *gorm.DB, c *fiber.Ctx) error {
	return nil
}

func ChatRoomExists(db *gorm.DB, user1ID, user2ID uint) (bool, error, uint) {
	RoomResult := new(RoomExistsResult)

	err := db.Table("chat_rooms").Joins("JOIN char_room_members on chat_room_members.chat_room_id = chat_rooms.id").Where("chat_room_members.user_id IN (?)", []uint{user1ID, user2ID}).Group("chat_rooms.id").Having("COUNT(DISTINCT chat_room_members.user_id) = ? ", 2).First(RoomResult).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil, 0
		}
		return false, err, 0
	}

	return true, nil, RoomResult.roomID
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

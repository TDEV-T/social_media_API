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
}

func (cr *ChatRoom) TableName() string {
	return "chat_room"
}

func (cm *ChatMessage) TableName() string {
	return "chat_message"
}

func CreateTable(db *gorm.DB, c *fiber.Ctx) error {
	return nil
}

func ChatRoomExists(db *gorm.DB, user1ID, user2ID uint) (bool, error) {
	var count int64

	err := db.Table("chat_rooms").Joins("JOIN char_room_members on chat_room_members.chat_room_id = chat_rooms.id").Where("chat_room_members.user_id IN (?)", []uint{user1ID, user2ID}).Group("chat_rooms.id").Having("COUNT(DISTINCT chat_room_members.user_id) = ? ", 2).Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

package models

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Chat_Room struct {
	gorm.Model
	SenderUserID   uint
	ReceiverUserID uint
	StatusChat     uint `gorm:"default:active"`
	sender         User `gorm:"foreignkey:SenderUserID"`
	receiver       User `gorm:"foreignkey:ReceiverUserID"`
}

type Chat_Message struct {
	ChatRoomID uint
	ChatRoom   Chat_Room `gorm:""`
}

func (ch *Chat_Room) TableName() string {
	return "chat_rooms"
}

func CreateTable(db *gorm.DB, c *fiber.Ctx) error {
	return nil
}

package models

import "gorm.io/gorm"

type Chat struct {
	gorm.Model
	SenderUserID   uint
	ReceiverUserID uint
	Message        string
}

func (ch *Chat) TableName() string {
	return "chats"
}

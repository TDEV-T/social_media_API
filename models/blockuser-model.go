package models

import "gorm.io/gorm"

type BlockedUser struct {
	gorm.Model
	BlockingUserID uint
	BlockedUserID  uint
}

func (bu *BlockedUser) TableName() string {
	return "blocked_users"
}

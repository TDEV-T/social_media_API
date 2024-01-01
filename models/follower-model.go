package models

import "gorm.io/gorm"

type Follower struct {
	gorm.Model
	FollowingUserID uint
	FollowerUserID  uint
}

func (f *Follower) TableName() string {
	return "followers"
}

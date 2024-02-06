package models

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type stats struct {
	User    int64
	Post    int64
	Like    int64
	Comment int64
}

func GetAllStats(db *gorm.DB, c *fiber.Ctx) error {
	stats := new(stats)

	var countUser int64
	db.Model(&User{}).Count(&countUser)

	var countPost int64
	db.Model(&Post{}).Count(&countPost)

	var countLike int64
	db.Model(&Like{}).Count(&countLike)

	var countComment int64
	db.Model(&Comment{}).Count(&countComment)

	stats.User = countUser
	stats.Comment = countComment
	stats.Post = countPost
	stats.Like = countLike

	return c.JSON(stats)
}

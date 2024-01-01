package models

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	UserID   uint
	Content  string    `json:"content"`
	Image    string    `gorm:"type:json"`
	IsPublic bool      `gorm:"default:true" json:"public"`
	Likes    []Like    `gorm:"foreignKey:PostID"`
	Comments []Comment `gorm:"foreignKey:PostID"`
}

func (p *Post) TableName() string {
	return "posts"
}

func CreatePost(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error Can't get UID"})
	}

	post := new(Post)

	if err := c.BodyParser(post); err != nil {
		return err
	}

	post.UserID = userLocal.ID

	imageLocal := c.Locals("images").([]string)

	imageJson, err := json.Marshal(imageLocal)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if len(imageLocal) != 0 {
		post.Image = string(imageJson)
	}

	fmt.Println(post)

	if post.Content == "" && post.Image == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Content And Image null",
		})
	}

	result := db.Save(post)

	if result.Error != nil {
		return result.Error
	}

	return c.JSON(fiber.Map{
		"message": "Create Post Success",
	})
}

func GetPosts(db *gorm.DB, c *fiber.Ctx) error {
	var PostsWithComment []Post

	result := db.Preload("Comments").Find(&PostsWithComment)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.JSON(PostsWithComment)

}

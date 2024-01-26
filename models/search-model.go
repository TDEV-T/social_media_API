package models

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Search struct {
	SearchText string `json:"search"`
}

func SearchFunc(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Can't get User Data"})
	}

	search := new(Search)

	if err := c.BodyParser(search); err != nil {
		return err
	}

	fmt.Println(search.SearchText)

	userFind, err := searchOnUser(db, search.SearchText, userLocal.ID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	postFind, err := searchOnPost(db, search.SearchText)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"userFind": userFind,
		"postFind": postFind,
	})
}

func searchOnUser(db *gorm.DB, search string, uid uint) ([]User, error) {

	var userList []User

	result := db.Table("users").Select("id", "username", "full_name", "profile_picture").Where("id != ? AND username LIKE ? OR full_name LIKE ?", uid, "%"+search+"%", "%"+search+"%").Find(&userList)

	if result.Error != nil {
		return nil, result.Error
	}

	return userList, nil
}

func searchOnPost(db *gorm.DB, search string) ([]Post, error) {
	var postList []Post
	result := db.Table("posts").Where("content LIKE ? ", "%"+search+"%").Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Comments").Preload("Comments.User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Likes").Preload("Likes.User", func(db *gorm.DB) *gorm.DB { return db.Select("id") }).Find(&postList).Find(&postList)

	if result.Error != nil {
		return nil, result.Error
	}

	return postList, nil
}

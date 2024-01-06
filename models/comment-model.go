package models

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	UserID  uint
	PostID  uint   `json:"pid"`
	Content string `json:"content"`
}

func (c *Comment) TableName() string {
	return "comments"
}

func CommentCreate(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error Can't get UID"})
	}

	comment := new(Comment)

	if err := c.BodyParser(comment); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	if userLocal.ID == 0 {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"message": "Can't Get UID"})
	}

	comment.UserID = userLocal.ID

	fmt.Println(comment)

	result := db.Save(comment)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString(result.Error.Error())
	}

	return c.JSON(fiber.Map{
		"message": "Create Comment Success",
	})

}

func CommentEdit(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error Can't get UID"})
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error Can't Get id comment"})
	}

	cm := Comment{Content: ""}

	if err := c.BodyParser(&cm); err != nil {
		return err
	}

	result := db.Model(&Comment{}).Where("id = ?", id).UpdateColumns(&cm)

	if result.Error != nil {
		return result.Error
	}

	return c.JSON(fiber.Map{
		"message": "Update Success",
	})
}

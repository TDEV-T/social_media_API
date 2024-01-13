package models

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Comment struct {
	gorm.Model
	UserID  uint
	PostID  uint   `json:"pid"`
	Content string `gorm:"not null" json:"content"`
	User    User   `gorm:"foreignkey:UserID"`
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

func DeleteComment(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error Can't Get Id"})
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	comment := new(Comment)

	result := db.First(comment, id)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": result.Error.Error()})
	}

	if comment.UserID != userLocal.ID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "U not owner this comment !"})
	}

	result = db.Unscoped().Delete(&Comment{}, id)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": result.Error.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Delete Comment Success"})
}

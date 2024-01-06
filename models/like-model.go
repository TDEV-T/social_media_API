package models

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Like struct {
	gorm.Model
	UserID uint
	PostID uint `json:"pid"`
}

func (l *Like) TableName() string {
	return "likes"
}

func LikePost(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Can't get user locasl"})
	}

	Like := new(Like)

	if err := c.BodyParser(Like); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
	}

	Like.UserID = userLocal.ID

	result := db.Save(Like)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": result.Error.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Success Like"})

}

func DeleteLike(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Can't get user locasl"})
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	result := db.Unscoped().Delete(&Like{}, id)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": result.Error.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Unlike Success"})

}

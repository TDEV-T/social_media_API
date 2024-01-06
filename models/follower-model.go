package models

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Follower struct {
	gorm.Model
	FollowingUserID uint
	FollowerUserID  uint `json:"follower"`
	Status          string
}

func (f *Follower) TableName() string {
	return "followers"
}

func RequestFollower(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error can't get User Local"})
	}

	follower := new(Follower)

	if err := c.BodyParser(follower); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	follower.FollowingUserID = userLocal.ID
	follower.Status = "pending"

	result := db.Save(follower)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "following success !"})

}

package models

import (
	"strconv"

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

	var existingFollower Follower

	result := db.Where("following_user_id = ? AND follower_user_id = ?", follower.FollowingUserID, follower.FollowerUserID).First(&existingFollower)

	if result.Error == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"message": "Follower request already exists"})
	}

	result = db.Create(follower)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "following success !"})

}

func AcceptFollower(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error can't get User Local"})
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error can't Get id "})
	}

	follower := Follower{Status: "accept"}

	result := db.Model(&Follower{}).Where("id = ?", id).UpdateColumns(follower)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("Can't find ")
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"message": "Accept Follow Request !"})

}

func UnFollower(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error can't get User Local"})
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error can't get id"})
	}

	result := db.Unscoped().Delete(&Follower{}, id)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": result.Error.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Delete Success",
	})

}

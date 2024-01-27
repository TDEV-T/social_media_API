package models

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Follower struct {
	gorm.Model
	FollowingUserID uint `json:"following"`
	FollowerUserID  uint `json:"follower"`
	Status          string
	Follower        User `gorm:"foreignKey:FollowerUserID"`
	Following       User `gorm:"foreignKey:FollowingUserID"`
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

	follower.FollowerUserID = userLocal.ID
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

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "following success !"})

}

func AcceptFollower(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error can't get User Local"})
	}

	follow := new(Follower)

	if err := c.BodyParser(follow); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error()})
	}

	follower := Follower{Status: "accept"}

	result := db.Model(&Follower{}).Where("follower_user_id = ? AND following_user_id = ?", userLocal.ID, follow.FollowingUserID).UpdateColumns(follower)

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

	following := new(Follower)

	if err := c.BodyParser(following); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error()})
	}

	result := db.Unscoped().Where("follower_user_id = ? AND following_user_id = ?", userLocal.ID, following.FollowingUserID).Delete(&Follower{})

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": result.Error.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Delete Success",
	})

}

func GetFollowingRequest(db *gorm.DB, userID uint) ([]Follower, error) {

	var FollowingRequest []Follower

	result := db.Where("following_user_id = ? ", userID).Preload("Follower", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Following", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Find(&FollowingRequest)

	if result.Error != nil {
		return nil, result.Error
	}

	return FollowingRequest, nil

}

func GetFollowerRequest(db *gorm.DB, userID uint) ([]Follower, error) {
	var FollowerRequest []Follower
	result := db.Where("follower_user_id = ?", userID).Preload("Follower", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Following", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Find(&FollowerRequest)

	if result.Error != nil {
		return nil, result.Error
	}

	return FollowerRequest, nil
}

func GetAllRequest(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Error can't get User Data"})
	}

	FollowerReq, err := GetFollowerRequest(db, userLocal.ID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	FollowingReq, err := GetFollowingRequest(db, userLocal.ID)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"FollowerRequest": FollowerReq, "FollowingRequset": FollowingReq})
}

func getFollowerAndFollowingCount(db *gorm.DB, userID uint) (int64, int64, error) {

	var followerCount int64
	var followingCount int64

	result := db.Model(&Follower{}).Where("following_user_id = ? AND status = 'active'", userID).Count(&followerCount)

	if result.Error != nil {
		return 0, 0, result.Error
	}

	result = db.Model(&Follower{}).Where("follower_user_id = ? AND status = 'active'", userID).Count(&followingCount)

	if result.Error != nil {
		return 0, 0, result.Error
	}

	return followerCount, followingCount, nil
}

func checkFollowRequest(db *gorm.DB, userID uint, otherUserID uint) (string, error) {

	if userID == otherUserID {
		return "myself", nil
	}

	Follow := new(Follower)

	result := db.Model(&Follower{}).Where("follower_user_id = ? AND following_user_id = ? ", userID, otherUserID).Find(Follow)
	if result.Error != nil {
		return "", result.Error
	}
	if Follow.Status != "" {
		if Follow.Status == "pending" {
			return "pending_send", nil
		} else if Follow.Status == "active" {
			return "active_send", nil
		}
	}

	result = db.Model(&Follower{}).Where("follower_user_id = ? AND following_user_id = ? ", otherUserID, userID).Find(Follow)

	if result.Error != nil {
		return "", result.Error
	}

	if Follow.Status != "" {
		if Follow.Status == "pending" {
			return "pending_recei", nil
		} else if Follow.Status == "active" {
			return "active_recei", nil
		}
	}

	return "none", nil
}

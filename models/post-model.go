package models

import (
	"encoding/json"
	"strconv"

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

	result := db.Preload("Comments").Preload("Likes").Find(&PostsWithComment)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.JSON(PostsWithComment)

}

func GetPostsPublic(db *gorm.DB) ([]Post, error) {
	var publicPost []Post

	result := db.Where("is_public = ?", true).Preload("Comments").Preload("Likes").Find(&publicPost)

	if result.Error != nil {
		return nil, result.Error
	}

	return publicPost, nil
}

func GetPostsFollower(db *gorm.DB, userId uint) ([]Post, error) {

	var friendPosts []Post

	friendIDs, err := GetFriendIDs(db, userId)

	if err != nil {
		return nil, err
	}

	if friendIDs == nil {
		return nil, nil
	}

	result := db.Where("user_id IN ?", friendIDs).Preload("Comments").Preload("Likes").Find(&friendPosts)

	if result.Error != nil {
		return nil, result.Error
	}

	return friendPosts, nil

}

func GetFeeds(db *gorm.DB, c *fiber.Ctx) error {

	userLocal := c.Locals("user").(*User)

	publicPosts, err := GetPostsPublic(db)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	friendPosts, err := GetPostsFollower(db, userLocal.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	allFeed := append(publicPosts, friendPosts...)

	if allFeed != nil && len(allFeed) > 0 {
		SortPostByCreatedAt(allFeed)
	}

	return c.JSON(allFeed)

}

func SortPostByCreatedAt(posts []Post) {

	if len(posts) < 2 {
		return
	}

	for i := 0; i < len(posts)-1; i++ {
		for j := 0; j < len(posts)-1-i; j++ {
			if posts[j].CreatedAt.Before(posts[j+1].CreatedAt) {
				posts[j], posts[j+1] = posts[j+1], posts[j]
			}
		}
	}
}

func DeletePosts(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	if userLocal.Role != "admin" {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	var post Post

	result := db.Delete(&post, id)

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).SendString(result.Error.Error())
	}

	return c.Status(fiber.StatusAccepted).SendString("Delete Success !")

}

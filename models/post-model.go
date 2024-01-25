package models

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Post struct {
	gorm.Model
	UserID      uint
	Content     string    `json:"content"`
	ContentType string    `gorm:"default:'picture'" json:"contenttype"`
	Image       string    `gorm:"type:json"`
	IsPublic    bool      `gorm:"default:true" json:"public"`
	Likes       []Like    `gorm:"foreignKey:PostID"`
	Comments    []Comment `gorm:"foreignKey:PostID"`
	User        User      `gorm:"foreignkey:UserID"`
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

	post.Image = string(imageJson)

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

	result := db.Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Comments").Preload("Comments.User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Likes").Preload("Likes.User", func(db *gorm.DB) *gorm.DB { return db.Select("id") }).Find(&PostsWithComment)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.JSON(PostsWithComment)

}

func GetPostById(db *gorm.DB, c *fiber.Ctx) error {

	id := c.Params("id")

	var PostsWithComment Post

	result := db.Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Comments").Preload("Comments.User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Likes").Preload("Likes.User", func(db *gorm.DB) *gorm.DB { return db.Select("id") }).Where("id = ?", id).Find(&PostsWithComment)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	return c.JSON(PostsWithComment)

}

func GetPostsPublic(db *gorm.DB) ([]Post, error) {
	var publicPost []Post

	result := db.Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Comments").Preload("Comments.User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Likes").Preload("Likes.User", func(db *gorm.DB) *gorm.DB { return db.Select("id") }).Find(&publicPost)

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

	// if friendIDs == nil {
	// 	return nil, nil
	// }

	result := db.Where("user_id IN ?", friendIDs).Preload("User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Comments").Preload("Comments.User", func(db *gorm.DB) *gorm.DB { return db.Select("id", "username", "full_name", "profile_picture") }).Preload("Likes").Preload("Likes.User", func(db *gorm.DB) *gorm.DB { return db.Select("id") }).Find(&friendPosts)
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

func GetFollowerFeed(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)

	followerPosts, err := GetPostsFollower(db, userLocal.ID)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	fmt.Println(followerPosts)

	return c.JSON(followerPosts)

}

func UpdatePost(db *gorm.DB, c *fiber.Ctx) error {
	userLocal := c.Locals("user").(*User)
	pid := c.Params("pid")
	imageCurrent := c.FormValue("imageCurrent")

	if userLocal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Can't get User Data",
		})
	}

	post := new(Post)

	db.First(&post, pid)

	if post.UserID != userLocal.ID {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Error Unauthorized",
		})
	}

	if err := c.BodyParser(post); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	imageLocal := c.Locals("images").([]string)
	var imageCurString []string

	if imageCurrent != "" {
		json.Unmarshal([]byte(imageCurrent), &imageCurString)

		if len(imageCurString) > 0 {
			imageLocal = append(imageLocal, imageCurString...)
		}
	}

	imageJson, err := json.Marshal(imageLocal)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	oldImage := post.Image
	post.Image = string(imageJson)

	if post.Content == "" && post.Image == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Content or Image Is Null",
		})
	}

	result := db.Save(post)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	var oldImageFiles []string

	json.Unmarshal([]byte(oldImage), &oldImageFiles)

	for _, oldImage := range oldImageFiles {
		if !contains(imageLocal, oldImage) {
			os.Remove("uploads/" + oldImage)
			fmt.Println("Delete File : " + oldImage)
		}
	}

	return c.JSON(fiber.Map{
		"message": "Update Post Success",
	})

}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false

}

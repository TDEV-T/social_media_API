package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"tdev/middleware"
	"tdev/models"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "root"
	password = "1329Pathrapol!"
	dbname   = "social_media"
)

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error Loading .env file")
	}

	portOpen := os.Getenv("PORT")

	if portOpen == "" {
		portOpen = "8080"
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})

	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.User{}, &models.Post{}, &models.Like{}, &models.Comment{}, &models.Follower{}, &models.BlockedUser{}, &models.Chat{})

	app := fiber.New()
	app.Post("/login", func(c *fiber.Ctx) error {
		return models.LoginUser(db, c)
	})
	app.Post("/register", func(c *fiber.Ctx) error {

		return models.CreateUser(db, c)
	})

	app.Use("/users", middleware.AuthRequired)

	app.Get("/users", func(c *fiber.Ctx) error {
		return c.JSON(models.GetUserAll(db, c))
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		return models.GetUserById(db, c)
	})

	app.Put("/users/:id", func(c *fiber.Ctx) error {
		return models.UpdateUser(db, c)
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return models.DeleteUser(db, c)
	})

	app.Use("/posts", middleware.AuthRequired)

	app.Post("/posts", middleware.UploadFile, func(c *fiber.Ctx) error {
		return models.CreatePost(db, c)
	})
	app.Get("/posts", func(c *fiber.Ctx) error {
		return models.GetPosts(db, c)
	})

	app.Use("/comment", middleware.AuthRequired)
	app.Post("/comment", func(c *fiber.Ctx) error {
		return models.CommentCreate(db, c)
	})

	app.Listen(":" + portOpen)
}

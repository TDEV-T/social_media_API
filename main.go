package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"tdev/functional"
	"tdev/middleware"
	"tdev/models"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	host     = "postgresChitter"
	port     = 5432
	user     = "root"
	password = "1329Pathrapol!"
	dbname   = "social_media"
)

var db *gorm.DB

func main() {

	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error Loading .env file")
	}

	portOpen := os.Getenv("PORT")

	if portOpen == "" {
		portOpen = "8080"
	}

	db = SetupDatabase()

	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&models.User{}, &models.Post{}, &models.Like{}, &models.Comment{}, &models.Follower{}, &models.BlockedUser{}, &models.ChatMessage{}, &models.ChatRoom{})

	app := fiber.New(fiber.Config{
		BodyLimit:         20 * 1024 * 1024,
		StreamRequestBody: true,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000,http://localhost:5890",
		AllowHeaders: "Origin,Content-Type,Accept,authtoken,file",
	}))

	setUpRoute(app)

	app.Listen(":" + portOpen)
}

func SetupDatabase() *gorm.DB {
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
		log.Fatal(err)
	}

	return db
}

var chatServer = &functional.ChatServer{
	Clients: make(map[*websocket.Conn]struct{}),
}

func setUpRoute(app *fiber.App) {

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendString("Test")
	})

	app.Post("/login", func(c *fiber.Ctx) error {
		return models.LoginUser(db, c)
	})
	app.Post("/register", func(c *fiber.Ctx) error {

		return models.CreateUser(db, c)
	})

	app.Post("/checkCurUser", middleware.AuthRequiredHeader, func(c *fiber.Ctx) error {
		return models.CheckCurUser(db, c)
	})

	app.Use("/users", middleware.AuthRequiredHeader)

	app.Post("/users/changepassword", func(c *fiber.Ctx) error {
		return models.UpdatePassword(db, c)
	})

	app.Get("/users/:id", func(c *fiber.Ctx) error {
		return models.GetUserById(db, c)
	})

	app.Put("/users/:id", middleware.UploadProfilePicture, func(c *fiber.Ctx) error {
		return models.UpdateUser(db, c)
	})

	app.Delete("/users/:id", func(c *fiber.Ctx) error {
		return models.DeleteUser(db, c)
	})

	app.Use("/posts", middleware.AuthRequiredHeader)

	app.Post("/posts", middleware.UploadFile, func(c *fiber.Ctx) error {
		return models.CreatePost(db, c)
	})
	app.Get("/posts", func(c *fiber.Ctx) error {
		return models.GetPosts(db, c)
	})

	app.Get("/posts/follower", func(c *fiber.Ctx) error {
		return models.GetFollowerFeed(db, c)
	})
	app.Get("/posts/feed", func(c *fiber.Ctx) error {
		return models.GetFeeds(db, c)
	})

	app.Get("/posts/:id", func(c *fiber.Ctx) error {
		return models.GetPostById(db, c)
	})
	app.Delete("/posts/:id", func(c *fiber.Ctx) error {
		return models.DeletePosts(db, c)
	})
	app.Patch("/posts/:pid", middleware.UploadFile, func(c *fiber.Ctx) error {
		return models.UpdatePost(db, c)
	})

	app.Use("/comment", middleware.AuthRequiredHeader)
	app.Post("/comment", func(c *fiber.Ctx) error {
		return models.CommentCreate(db, c)
	})
	app.Patch("/comment/:id", func(c *fiber.Ctx) error {
		return models.CommentEdit(db, c)
	})
	app.Delete("/comment/:id", func(c *fiber.Ctx) error {
		return models.DeleteComment(db, c)
	})

	app.Use("/follow", middleware.AuthRequiredHeader)
	app.Get("/follow", func(c *fiber.Ctx) error {
		return models.GetAllRequest(db, c)
	})
	app.Post("/follow/req", func(c *fiber.Ctx) error {
		return models.RequestFollower(db, c)
	})
	app.Post("/follow/accept/:id", func(c *fiber.Ctx) error {
		return models.AcceptFollower(db, c)
	})
	app.Delete("/follow/unfollow", func(c *fiber.Ctx) error {
		return models.UnFollower(db, c)
	})
	app.Delete("/follow/reject", func(c *fiber.Ctx) error {
		return models.RejectFollow(db, c)
	})

	app.Use("/like", middleware.AuthRequiredHeader)
	app.Post("/like", func(c *fiber.Ctx) error {
		return models.LikePost(db, c)
	})
	app.Post("/like/unlike/:id", func(c *fiber.Ctx) error {
		return models.DeleteLike(db, c)
	})

	app.Post("/uploadfile", middleware.UploadFile, func(c *fiber.Ctx) error {
		return functional.UploadFile(c)
	})

	app.Delete("/uploadfile/:name", func(c *fiber.Ctx) error {
		return functional.DeleteFile(c)
	})

	app.Use("/search", middleware.AuthRequiredHeader)
	app.Post("/search", func(c *fiber.Ctx) error {
		return models.SearchFunc(db, c)
	})

	app.Get("/images/:imageName", functional.GetImageHandler)
	app.Get("/video/:video", functional.StreamVideo)

	app.Use("/chat", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Use("/chatAll", middleware.AuthRequiredQueryWithChat)
	app.Get("/chatAll", websocket.New(functional.GetAllChatRoomWithUserID(db)))

	app.Use("/chat", middleware.AuthRequiredHeaderForChat)
	app.Get("/chat", websocket.New(functional.MessageSocket(db, chatServer)))

	app.Use("/admin", middleware.AuthAdminRequireWithHeader)

	app.Get("/admin/users", func(c *fiber.Ctx) error {
		return c.JSON(models.GetUserAll(db, c))
	})

	app.Get("/admin/users/:id", func(c *fiber.Ctx) error {
		return models.GetUserById(db, c)
	})

	app.Get("/admin/posts", func(c *fiber.Ctx) error {
		return models.GetPosts(db, c)
	})

	app.Get("admin/stats", func(c *fiber.Ctx) error {
		return models.GetAllStats(db, c)
	})

	app.Post("/loginAdmin", func(c *fiber.Ctx) error {
		return models.LoginAdmin(db, c)
	})

}

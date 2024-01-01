package models

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username       string        `gorm:"unique;not null" json:"username"`
	Password       string        `gorm:"not null" json:"password"`
	FullName       string        `json:"fullname"`
	Email          string        `gorm:"unique" json:"email"`
	ProfilePicture string        `json:"profilepicture"`
	Bio            string        `json:"bio"`
	PrivateAccount bool          `gorm:"default:false"`
	Role           string        `gorm:"default:'user'"`
	Posts          []Post        `gorm:"foreignKey:UserID"`
	Likes          []Like        `gorm:"foreignKey:UserID"`
	Comments       []Comment     `gorm:"foreignKey:UserID"`
	Followers      []Follower    `gorm:"foreignKey:FollowingUserID"`
	BlockedUsers   []BlockedUser `gorm:"foreignKey:BlockingUserID"`
	SentChats      []Chat        `gorm:"foreignKey:SenderUserID"`
	ReceivedChats  []Chat        `gorm:"foreignKey:ReceiverUserID"`
}

func (u *User) TableName() string {
	return "users"
}

func CreateUser(db *gorm.DB, c *fiber.Ctx) error {

	user := new(User)

	if err := c.BodyParser(user); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)

	result := db.Create(&user)

	if result.Error != nil {
		// log.Fatal("Error Register User : %v", result.Error)
		return result.Error
	}

	return c.JSON(fiber.Map{
		"message": "Register Successfully",
	})
}

func GetUserById(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")

	var usr User
	result := db.First(&usr, id)

	if result.Error != nil {
		return result.Error
	}

	return c.JSON(&usr)
}

func UpdateUser(db *gorm.DB, c *fiber.Ctx) error {

	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	usr := new(User)
	if err := c.BodyParser(usr); err != nil {
		return err
	}

	usr.ID = uint(id)

	result := db.Save(&usr)

	if result.Error != nil {
		return result.Error
	}

	return c.JSON(fiber.Map{
		"message": "Update User Success",
	})

}

func DeleteUser(db *gorm.DB, c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))

	if err != nil {
		return c.SendStatus(fiber.StatusBadRequest)
	}
	usr := new(User)
	result := db.Delete(&usr, id)

	if result.Error != nil {
		return result.Error
	}

	return c.JSON(fiber.Map{
		"message": "Delete Success",
	})

}

func FindUser(db *gorm.DB, username string) []User {
	var users []User

	result := db.Where("username LIKE ?", username).Order("created_at desc").Find(&users)

	if result.Error != nil {
		log.Fatalf("Search User failed : %v", result.Error)
	}

	return users

}

func GetUserAll(db *gorm.DB, c *fiber.Ctx) []User {

	userLocal := c.Locals("user").(*User)

	fmt.Printf("User : %s , Role : %s", userLocal.Username, userLocal.Role)

	var user []User

	result := db.Find(&user)

	if result.Error != nil {
		log.Fatalf("Find User failed : %v", result.Error)
	}

	return user
}

func LoginUser(db *gorm.DB, c *fiber.Ctx) error {

	user := new(User)

	selectedUser := new(User)

	if err := c.BodyParser(user); err != nil {
		return err
	}

	result := db.Where("username = ?", user.Username).First(selectedUser)

	if result.Error != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(selectedUser.Password), []byte(user.Password)); err != nil {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	jwtSecretKey := os.Getenv("SECRET_KEY")

	if jwtSecretKey == "" {
		return c.JSON(fiber.Map{
			"message": "Can't get secret_key",
		})
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = selectedUser.ID
	claims["user_name"] = selectedUser.Username
	claims["user_role"] = selectedUser.Role
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	t, err := token.SignedString([]byte(jwtSecretKey))

	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    t,
		Expires:  time.Now().Add(time.Hour * 72),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"token":    t,
		"userid":   selectedUser.ID,
		"username": selectedUser.Username,
		"userrole": selectedUser.Role,
		"message":  "Login Successfully",
	})

}

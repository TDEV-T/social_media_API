package middleware

import (
	"os"
	"tdev/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
)

const userContextKey = "user"

func AuthRequired(c *fiber.Ctx) error {

	user := new(models.User)
	cookie := c.Cookies("jwt")
	jwtSecretKey := os.Getenv("SECRET_KEY")

	if jwtSecretKey == "" {
		return c.JSON(fiber.Map{"message": "Can't get jwt secret"})
	}

	token, err := jwt.ParseWithClaims(cookie, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecretKey), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}

	claim := token.Claims.(jwt.MapClaims)

	username, ok := claim["user_name"].(string)

	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	role, ok := claim["user_role"].(string)

	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	uidFloat, ok := claim["user_id"].(float64)

	if !ok {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(err.Error())
	}

	uid := int(uidFloat)

	user.ID = uint(uid)
	user.Username = username
	user.Role = role

	c.Locals(userContextKey, user)

	return c.Next()
}

package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

func UploadFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var filenames []string

	for _, files := range form.File {
		for _, file := range files {
			timestamp := time.Now().Unix()
			newFileName := fmt.Sprintf("%d_%s", timestamp, file.Filename)

			destination := "./uploads/" + newFileName
			err = c.SaveFile(file, destination)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
			}

			filenames = append(filenames, newFileName)
		}
	}

	c.Locals("images", filenames)

	return c.Next()
}

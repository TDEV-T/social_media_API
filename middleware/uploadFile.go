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

func UploadProfilePicture(c *fiber.Ctx) error {
	fileProfile, err := c.FormFile("profilepicture")
	if err != nil || fileProfile == nil {
		oldImage := c.FormValue("oldImage")
		c.Locals("profilepicture", oldImage)
	} else {
		newProfileName := createNewFileName(fileProfile.Filename)
		destination := "./uploads/" + newProfileName

		if err := c.SaveFile(fileProfile, destination); err != nil {
			return err
		}

		c.Locals("profilepicture", newProfileName)
	}

	fileCloverPicture, err := c.FormFile("coverpicture")
	if err != nil || fileCloverPicture == nil {
		oldClover := c.FormValue("oldClover")
		c.Locals("cloverpicture", oldClover)
	} else {
		newFileClover := createNewFileName(fileCloverPicture.Filename)
		destination2 := "./uploads/" + newFileClover

		if err := c.SaveFile(fileCloverPicture, destination2); err != nil {
			return err
		}

		c.Locals("cloverpicture", newFileClover)
	}

	return c.Next()
}

func createNewFileName(oldFileName string) string {
	timestamp := time.Now().Unix()
	newFileName := fmt.Sprintf("%d_%s", timestamp, oldFileName)

	return newFileName
}

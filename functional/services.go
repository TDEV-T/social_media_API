package functional

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func GetRootPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	return exPath
}

func DeleteFile(c *fiber.Ctx) error {
	filename := c.Params("name")

	if filename == "" {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	filePath := filepath.Join("uploads", filename)

	fmt.Println(filePath)

	err := os.Remove(filePath)
	if err != nil {
		fmt.Println(err.Error())
		return c.SendStatus(fiber.StatusInternalServerError)
	}
	return c.SendStatus(fiber.StatusOK)
}

func UploadFile(c *fiber.Ctx) error {
	imageLocal := c.Locals("images").([]string)

	imageJson, err := json.Marshal(imageLocal)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Error Upload file"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Upload Successfully !", "data": imageLocal, "json": imageJson})
}

func GetImageHandler(c *fiber.Ctx) error {
	imageName := c.Params("imageName")

	imagePath := filepath.Join("uploads", imageName)

	imageData, err := ioutil.ReadFile(imagePath)

	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	return c.Send(imageData)
}

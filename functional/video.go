package functional

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func StreamVideo(c *fiber.Ctx) error {
	videoName := c.Params("video")
	videoPath := filepath.Join("uploads", videoName)

	videoFile, err := os.Open(videoPath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Video not found")
	}
	defer videoFile.Close()

	c.Set("Content-Type", "video/mp4")
	c.Set("Accept-Ranges", "bytes")

	rangeHeader := c.Get("Range")
	if rangeHeader != "" {
		return streamRangeVideo(c, videoFile, rangeHeader)
	}

	return c.Status(http.StatusOK).SendStream(videoFile)
}

func streamRangeVideo(c *fiber.Ctx, videoFile *os.File, rangeHeader string) error {

	size, err := videoFile.Stat()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Cannot access video file")
	}
	start, length, end, err := parseRange(rangeHeader, size.Size())
	if err != nil {
		return c.Status(fiber.StatusRequestedRangeNotSatisfiable).SendString("Invalid range")
	}

	startStr := strconv.FormatInt(start, 10)
	endStr := strconv.FormatInt(start, 10)
	sizeStr := strconv.FormatInt(size.Size(), 10)

	c.Set("Content-Range", "bytes "+startStr+"-"+endStr+"/"+sizeStr)
	c.Set("Content-Length", strconv.FormatInt(length, 10))
	c.Set("Content-Type", "video/mp4")
	c.Status(fiber.StatusPartialContent)

	_, err = videoFile.Seek(start, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Cannot seek video file")
	}
	reader := io.NewSectionReader(videoFile, start, end-start+1)

	return c.SendStream(reader)
}

func parseRange(rangeHeader string, size int64) (start, end, length int64, err error) {
	// ตัวอย่าง Range header: "bytes=0-499"
	if rangeHeader == "" {
		// ไม่มี range header ส่งค่าเริ่มต้นและสิ้นสุดเต็มขนาดไฟล์
		return 0, size - 1, size, nil
	}

	// ลบคำว่า "bytes=" และแยกส่วนตัวเลข
	rangeParts := strings.Split(rangeHeader, "=")
	if len(rangeParts) != 2 {
		return 0, 0, 0, fmt.Errorf("invalid range header")
	}

	// แยกส่วนตัวเลขเริ่มต้นและสิ้นสุด
	rangeStartEnd := strings.Split(rangeParts[1], "-")
	if len(rangeStartEnd) != 2 {
		return 0, 0, 0, fmt.Errorf("invalid range header")
	}

	// แปลงสตริงเป็นตัวเลข int64
	start, err = strconv.ParseInt(rangeStartEnd[0], 10, 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid start range")
	}

	end, err = strconv.ParseInt(rangeStartEnd[1], 10, 64)
	if err != nil {
		end = size - 1 // ถ้าไม่มี end range ให้ใช้ขนาดไฟล์สุดท้าย
	}

	length = end - start + 1 // คำนวณความยาวของช่วง

	// ตรวจสอบความถูกต้องของช่วง
	if start < 0 || end >= size || start > end {
		return 0, 0, 0, fmt.Errorf("invalid range")
	}

	return start, end, length, nil
}

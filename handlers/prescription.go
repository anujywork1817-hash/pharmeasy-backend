package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadPrescription(c *gin.Context) {
	userID := c.GetUint("user_id")
	fmt.Println("📋 Uploading prescription for user_id:", userID)

	// Get file from request
	file, err := c.FormFile("prescription")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".pdf":  true,
	}

	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file type. Only JPG, PNG, PDF allowed",
		})
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File too large. Max size is 5MB",
		})
		return
	}

	// Create uploads directory if not exists
	uploadDir := "./uploads/prescriptions"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	fileName := fmt.Sprintf("prescription_%d_%d%s",
		userID,
		time.Now().Unix(),
		ext,
	)
	filePath := filepath.Join(uploadDir, fileName)

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Return file URL
	fileURL := fmt.Sprintf("/uploads/prescriptions/%s", fileName)
	fmt.Println("✅ Prescription uploaded:", fileURL)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Prescription uploaded successfully",
		"file_url": fileURL,
		"filename": fileName,
	})
}

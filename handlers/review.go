package handlers

import (
	"fmt"
	"net/http"

	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

func AddReview(c *gin.Context) {
	userID := c.GetUint("user_id")
	fmt.Println("⭐ AddReview called by user:", userID)

	var input struct {
		MedicineID uint    `json:"medicine_id" binding:"required"`
		Rating     float64 `json:"rating" binding:"required,min=1,max=5"`
		Comment    string  `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("❌ Bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("📦 Review input:", input)

	// Check if already reviewed
	var existing models.Review
	err := config.DB.Where(
		"user_id = ? AND medicine_id = ?",
		userID, input.MedicineID,
	).First(&existing).Error

	if err == nil {
		// Update
		config.DB.Model(&existing).Updates(map[string]interface{}{
			"rating":  input.Rating,
			"comment": input.Comment,
		})
		fmt.Println("✅ Review updated:", existing.ID)
		c.JSON(http.StatusOK, gin.H{
			"message": "Review updated successfully",
			"review":  existing,
		})
		return
	}

	// Create new
	review := models.Review{
		UserID:     userID,
		MedicineID: input.MedicineID,
		Rating:     input.Rating,
		Comment:    input.Comment,
	}

	if err := config.DB.Create(&review).Error; err != nil {
		fmt.Println("❌ Create review error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review"})
		return
	}

	fmt.Println("✅ Review created:", review.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Review added successfully",
		"review":  review,
	})
}

func GetReviews(c *gin.Context) {
	medicineID := c.Param("id")
	fmt.Println("📋 GetReviews for medicine:", medicineID)

	var reviews []models.Review
	config.DB.
		Preload("User").
		Where("medicine_id = ?", medicineID).
		Order("created_at desc").
		Find(&reviews)

	var avgRating float64
	if len(reviews) > 0 {
		var total float64
		for _, r := range reviews {
			total += r.Rating
		}
		avgRating = total / float64(len(reviews))
	}

	fmt.Println("✅ Reviews found:", len(reviews), "avg:", avgRating)

	c.JSON(http.StatusOK, gin.H{
		"reviews":    reviews,
		"avg_rating": avgRating,
		"total":      len(reviews),
	})
}

func GetMyReview(c *gin.Context) {
	userID := c.GetUint("user_id")
	medicineID := c.Param("id")

	fmt.Println("👤 GetMyReview user:", userID, "medicine:", medicineID)

	var review models.Review
	if err := config.DB.Where(
		"user_id = ? AND medicine_id = ?",
		userID, medicineID,
	).First(&review).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"review": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"review": review})
}

func DeleteReview(c *gin.Context) {
	userID := c.GetUint("user_id")
	medicineID := c.Param("id")

	fmt.Println("🗑️ DeleteReview user:", userID, "medicine:", medicineID)

	result := config.DB.Where(
		"user_id = ? AND medicine_id = ?",
		userID, medicineID,
	).Delete(&models.Review{})

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Review deleted"})
}

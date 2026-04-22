package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

// Validate and apply coupon
func ValidateCoupon(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input struct {
		Code        string  `json:"code" binding:"required"`
		OrderAmount float64 `json:"order_amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := strings.ToUpper(strings.TrimSpace(input.Code))
	fmt.Println("🎟️ Validating coupon:", code, "for user:", userID)

	// Find coupon
	var coupon models.Coupon
	if err := config.DB.Where("code = ? AND is_active = true", code).
		First(&coupon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid coupon code"})
		return
	}

	// Check expiry
	if coupon.ExpiresAt != "" {
		expiry, err := time.Parse("2006-01-02", coupon.ExpiresAt)
		if err == nil && time.Now().After(expiry) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon has expired"})
			return
		}
	}

	// Check usage limit
	if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon usage limit reached"})
		return
	}

	// Check min order amount
	if input.OrderAmount < coupon.MinOrderAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"Minimum order amount is ₹%.0f for this coupon",
				coupon.MinOrderAmount,
			),
		})
		return
	}

	// Check if user already used this coupon
	var existingUsage models.CouponUsage
	if err := config.DB.Where(
		"coupon_id = ? AND user_id = ?",
		coupon.ID, userID,
	).First(&existingUsage).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You have already used this coupon",
		})
		return
	}

	// Calculate discount
	var discountAmount float64
	if coupon.DiscountType == "percent" {
		discountAmount = input.OrderAmount * coupon.DiscountValue / 100
		if coupon.MaxDiscount > 0 && discountAmount > coupon.MaxDiscount {
			discountAmount = coupon.MaxDiscount
		}
	} else {
		discountAmount = coupon.DiscountValue
	}

	// Ensure discount doesn't exceed order amount
	if discountAmount > input.OrderAmount {
		discountAmount = input.OrderAmount
	}

	finalAmount := input.OrderAmount - discountAmount

	fmt.Println("✅ Coupon valid! Discount:", discountAmount)

	c.JSON(http.StatusOK, gin.H{
		"valid":           true,
		"coupon":          coupon,
		"discount_amount": discountAmount,
		"final_amount":    finalAmount,
		"message": fmt.Sprintf(
			"Coupon applied! You save ₹%.0f",
			discountAmount,
		),
	})
}

// Get all active coupons
func GetActiveCoupons(c *gin.Context) {
	var coupons []models.Coupon
	config.DB.Where("is_active = true").Find(&coupons)
	fmt.Println("🎟️ Active coupons:", len(coupons))
	c.JSON(http.StatusOK, coupons)
}

// Create coupon (admin)
func CreateCoupon(c *gin.Context) {
	var input struct {
		Code           string  `json:"code" binding:"required"`
		Description    string  `json:"description"`
		DiscountType   string  `json:"discount_type" binding:"required"`
		DiscountValue  float64 `json:"discount_value" binding:"required"`
		MinOrderAmount float64 `json:"min_order_amount"`
		MaxDiscount    float64 `json:"max_discount"`
		UsageLimit     int     `json:"usage_limit"`
		ExpiresAt      string  `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupon := models.Coupon{
		Code:           strings.ToUpper(input.Code),
		Description:    input.Description,
		DiscountType:   input.DiscountType,
		DiscountValue:  input.DiscountValue,
		MinOrderAmount: input.MinOrderAmount,
		MaxDiscount:    input.MaxDiscount,
		UsageLimit:     input.UsageLimit,
		IsActive:       true,
		ExpiresAt:      input.ExpiresAt,
	}

	if err := config.DB.Create(&coupon).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Coupon code already exists"})
		return
	}

	fmt.Println("✅ Coupon created:", coupon.Code)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Coupon created successfully",
		"coupon":  coupon,
	})
}

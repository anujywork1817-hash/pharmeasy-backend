package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

func GetMedicines(c *gin.Context) {
	var medicines []models.Medicine
	search := c.Query("search")
	category := c.Query("category")

	fmt.Println("🔍 search:", search, "| category:", category)

	tx := config.DB

	if search != "" {
		tx = tx.Where(
			"name ILIKE ? OR brand ILIKE ? OR description ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%",
		)
	}

	if category != "" && category != "All" {
		tx = tx.Where("category = ?", category)
	}

	if err := tx.Find(&medicines).Error; err != nil {
		fmt.Println("❌ Query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("✅ Medicines returned:", len(medicines))
	c.JSON(http.StatusOK, medicines)
}

func GetMedicineByID(c *gin.Context) {
	id := c.Param("id")
	var medicine models.Medicine

	if err := config.DB.First(&medicine, id).Error; err != nil {
		fmt.Println("❌ Medicine not found, id:", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "Medicine not found"})
		return
	}

	fmt.Println("✅ Medicine found:", medicine.Name)
	c.JSON(http.StatusOK, medicine)
}

func PlaceOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	fmt.Println("🛒 Placing order for user_id:", userID)

	var input struct {
		Address    string `json:"address" binding:"required"`
		CouponCode string `json:"coupon_code"`
		Items      []struct {
			MedicineID uint `json:"medicine_id" binding:"required"`
			Quantity   int  `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var total float64
	var orderItems []models.OrderItem

	for _, item := range input.Items {
		var med models.Medicine
		if err := config.DB.First(&med, item.MedicineID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error":       "Medicine not found",
				"medicine_id": item.MedicineID,
			})
			return
		}

		if med.Stock < item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":    "Insufficient stock",
				"medicine": med.Name,
			})
			return
		}

		discountedPrice := med.Price
		if med.Discount > 0 {
			discountedPrice = med.Price * (1 - med.Discount/100)
		}
		total += discountedPrice * float64(item.Quantity)

		orderItems = append(orderItems, models.OrderItem{
			MedicineID: item.MedicineID,
			Quantity:   item.Quantity,
			Price:      discountedPrice,
		})

		config.DB.Model(&med).Update("stock", med.Stock-item.Quantity)
		fmt.Println("📦 Added item:", med.Name, "qty:", item.Quantity)
	}

	// Apply coupon if provided
	var discountAmount float64
	couponCode := strings.ToUpper(strings.TrimSpace(input.CouponCode))

	if couponCode != "" {
		var coupon models.Coupon
		if err := config.DB.Where(
			"code = ? AND is_active = true", couponCode,
		).First(&coupon).Error; err == nil {
			if coupon.DiscountType == "percent" {
				discountAmount = total * coupon.DiscountValue / 100
				if coupon.MaxDiscount > 0 &&
					discountAmount > coupon.MaxDiscount {
					discountAmount = coupon.MaxDiscount
				}
			} else {
				discountAmount = coupon.DiscountValue
			}

			if discountAmount > total {
				discountAmount = total
			}

			// Record usage
			config.DB.Create(&models.CouponUsage{
				CouponID: coupon.ID,
				UserID:   userID,
				Discount: discountAmount,
			})

			// Increment used count
			config.DB.Model(&coupon).Update(
				"used_count", coupon.UsedCount+1,
			)

			fmt.Println("🎟️ Coupon applied:", couponCode,
				"discount:", discountAmount)
		}
	}

	finalPrice := total - discountAmount

	order := models.Order{
		UserID:         userID,
		Address:        input.Address,
		Status:         "pending",
		TotalPrice:     total,
		DiscountAmount: discountAmount,
		FinalPrice:     finalPrice,
		CouponCode:     couponCode,
		Items:          orderItems,
	}

	if err := config.DB.Create(&order).Error; err != nil {
		fmt.Println("❌ Order creation error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to place order",
		})
		return
	}

	fmt.Println("✅ Order placed! ID:", order.ID,
		"Total:", total,
		"Discount:", discountAmount,
		"Final:", finalPrice)

	c.JSON(http.StatusCreated, gin.H{
		"message":         "Order placed successfully!",
		"order":           order,
		"discount_amount": discountAmount,
		"final_price":     finalPrice,
	})
}

// ✅ GetMyOrders — was missing
func GetMyOrders(c *gin.Context) {
	userID := c.GetUint("user_id")
	fmt.Println("📋 Fetching orders for user_id:", userID)

	var orders []models.Order
	result := config.DB.
		Preload("Items.Medicine").
		Preload("TrackingUpdates").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&orders)

	fmt.Println("📦 Orders found:", result.RowsAffected)
	c.JSON(http.StatusOK, orders)
}

// ✅ GetOrderByID — was missing
func GetOrderByID(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	fmt.Println("🔍 Fetching order id:", id, "for user_id:", userID)

	var order models.Order
	if err := config.DB.
		Preload("Items.Medicine").
		Preload("TrackingUpdates").
		Where("id = ? AND user_id = ?", id, userID).
		First(&order).Error; err != nil {
		fmt.Println("❌ Order not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	fmt.Println("✅ Order found:", order.ID)
	c.JSON(http.StatusOK, order)
}

// ✅ Cancel order
func CancelOrder(c *gin.Context) {
	userID := c.GetUint("user_id")
	id := c.Param("id")

	fmt.Println("❌ Cancelling order:", id, "for user:", userID)

	var order models.Order
	if err := config.DB.
		Where("id = ? AND user_id = ?", id, userID).
		First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Only pending orders can be cancelled
	if order.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only pending orders can be cancelled",
		})
		return
	}

	// Update status
	config.DB.Model(&order).Update("status", "cancelled")

	// Add tracking update
	config.DB.Create(&models.TrackingUpdate{
		OrderID:     order.ID,
		Status:      "cancelled",
		Description: "Order cancelled by user",
		Location:    "PharmEasy Store",
	})

	// Restore stock
	var items []models.OrderItem
	config.DB.Where("order_id = ?", order.ID).Find(&items)
	for _, item := range items {
		var med models.Medicine
		if err := config.DB.First(&med, item.MedicineID).Error; err == nil {
			config.DB.Model(&med).Update("stock", med.Stock+item.Quantity)
			fmt.Println("📦 Restored stock for:", med.Name)
		}
	}

	fmt.Println("✅ Order cancelled:", order.ID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
	})
}

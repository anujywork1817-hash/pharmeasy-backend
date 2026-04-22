package handlers

import (
	"fmt"
	"net/http"

	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

// Get order with full tracking
func GetOrderTracking(c *gin.Context) {
	userID := c.GetUint("user_id")
	orderID := c.Param("id")

	fmt.Println("📦 GetOrderTracking order:", orderID, "user:", userID)

	// First find the order without preloads
	var order models.Order
	if err := config.DB.
		Where("id = ? AND user_id = ?", orderID, userID).
		First(&order).Error; err != nil {
		fmt.Println("❌ Order not found:", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Load items separately
	config.DB.Preload("Medicine").Find(&order.Items, "order_id = ?", order.ID)

	// Load tracking updates separately
	config.DB.Where("order_id = ?", order.ID).
		Order("created_at asc").
		Find(&order.TrackingUpdates)

	// If no tracking updates exist, create default ones
	if len(order.TrackingUpdates) == 0 {
		createDefaultTracking(order)
		config.DB.Where("order_id = ?", order.ID).
			Order("created_at asc").
			Find(&order.TrackingUpdates)
	}

	fmt.Println("✅ Order found, tracking updates:", len(order.TrackingUpdates))
	c.JSON(http.StatusOK, order)
}

// Update order status (admin use)
func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	var input struct {
		Status      string `json:"status" binding:"required"`
		Description string `json:"description"`
		Location    string `json:"location"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Update order status
	config.DB.Model(&order).Update("status", input.Status)

	// Add tracking update
	description := input.Description
	if description == "" {
		description = getDefaultDescription(input.Status)
	}

	location := input.Location
	if location == "" {
		location = "Mumbai Warehouse"
	}

	tracking := models.TrackingUpdate{
		OrderID:     order.ID,
		Status:      input.Status,
		Description: description,
		Location:    location,
	}
	config.DB.Create(&tracking)

	fmt.Println("✅ Order status updated:", order.ID, "->", input.Status)

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated",
		"order":   order,
	})
}

func createDefaultTracking(order models.Order) {
	statuses := []struct {
		status      string
		description string
		location    string
	}{
		{
			"pending",
			"Order placed successfully",
			"PharmEasy Store",
		},
	}

	if order.Status == "confirmed" ||
		order.Status == "shipped" ||
		order.Status == "delivered" {
		statuses = append(statuses, struct {
			status      string
			description string
			location    string
		}{
			"confirmed",
			"Order confirmed and being prepared",
			"PharmEasy Warehouse, Mumbai",
		})
	}

	if order.Status == "shipped" ||
		order.Status == "delivered" {
		statuses = append(statuses, struct {
			status      string
			description string
			location    string
		}{
			"shipped",
			"Order picked up by delivery partner",
			"Mumbai Distribution Center",
		})
	}

	if order.Status == "delivered" {
		statuses = append(statuses, struct {
			status      string
			description string
			location    string
		}{
			"delivered",
			"Order delivered successfully",
			order.Address,
		})
	}

	for _, s := range statuses {
		tracking := models.TrackingUpdate{
			OrderID:     order.ID,
			Status:      s.status,
			Description: s.description,
			Location:    s.location,
		}
		config.DB.Create(&tracking)
	}
}

func getDefaultDescription(status string) string {
	switch status {
	case "pending":
		return "Order placed successfully"
	case "confirmed":
		return "Order confirmed and being prepared"
	case "shipped":
		return "Order picked up by delivery partner"
	case "out_for_delivery":
		return "Order is out for delivery"
	case "delivered":
		return "Order delivered successfully"
	case "cancelled":
		return "Order has been cancelled"
	default:
		return "Order status updated"
	}
}

// Simulate auto progression for testing
func SimulateTracking(c *gin.Context) {
	orderID := c.Param("id")

	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Progress to next status
	nextStatus := map[string]string{
		"pending":          "confirmed",
		"confirmed":        "shipped",
		"shipped":          "out_for_delivery",
		"out_for_delivery": "delivered",
	}

	next, ok := nextStatus[order.Status]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot progress from status: " + order.Status,
		})
		return
	}

	config.DB.Model(&order).Update("status", next)

	tracking := models.TrackingUpdate{
		OrderID:     order.ID,
		Status:      next,
		Description: getDefaultDescription(next),
		Location:    "Mumbai, Maharashtra",
	}
	config.DB.Create(&tracking)

	fmt.Println("✅ Simulated tracking:", order.ID, "->", next)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Order progressed",
		"new_status": next,
	})
}

package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

const (
	razorpayKeyID     = "rzp_test_Serg6vvG3Ygu8y"
	razorpayKeySecret = "D015QD3AJ4VtMwp5sEZ5McwP"
)

func CreatePaymentOrder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input struct {
		Amount int `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("💳 Creating Razorpay order for user:", userID, "amount:", input.Amount)
	fmt.Println("🔑 Key ID:", razorpayKeyID)

	orderData := map[string]interface{}{
		"amount":          input.Amount,
		"currency":        "INR",
		"receipt":         fmt.Sprintf("receipt_user_%d", userID),
		"partial_payment": false,
	}

	jsonData, _ := json.Marshal(orderData)

	req, err := http.NewRequest(
		"POST",
		"https://api.razorpay.com/v1/orders",
		strings.NewReader(string(jsonData)),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	req.SetBasicAuth(razorpayKeyID, razorpayKeySecret)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Razorpay API error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact Razorpay"})
		return
	}
	defer resp.Body.Close()

	var razorpayOrder map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&razorpayOrder)
	fmt.Println("📬 Razorpay response:", razorpayOrder)

	if razorpayOrder["error"] != nil {
		errMap := razorpayOrder["error"].(map[string]interface{})
		c.JSON(http.StatusBadRequest, gin.H{"error": errMap["description"]})
		return
	}

	fmt.Println("✅ Razorpay order created:", razorpayOrder["id"])

	c.JSON(http.StatusOK, gin.H{
		"order_id": razorpayOrder["id"],
		"amount":   razorpayOrder["amount"],
		"currency": razorpayOrder["currency"],
		"key_id":   razorpayKeyID,
	})
}

func VerifyPayment(c *gin.Context) {
	var input struct {
		RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
		RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
		RazorpaySignature string `json:"razorpay_signature" binding:"required"`
		Address           string `json:"address" binding:"required"`
		Items             []struct {
			MedicineID uint `json:"medicine_id"`
			Quantity   int  `json:"quantity"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify signature
	data := input.RazorpayOrderID + "|" + input.RazorpayPaymentID
	h := hmac.New(sha256.New, []byte(razorpayKeySecret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	fmt.Println("🔐 Signature match:", expectedSignature == input.RazorpaySignature)

	if expectedSignature != input.RazorpaySignature {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment verification failed"})
		return
	}

	PlaceOrderAfterPayment(c, input.Address, input.Items, input.RazorpayPaymentID)
}

func PlaceOrderAfterPayment(
	c *gin.Context,
	address string,
	items []struct {
		MedicineID uint `json:"medicine_id"`
		Quantity   int  `json:"quantity"`
	},
	paymentID string,
) {
	userID := c.GetUint("user_id")
	fmt.Println("📦 Placing order after payment for user:", userID)

	var orderItems []models.OrderItem
	var total float64

	for _, item := range items {
		var med models.Medicine
		if err := config.DB.First(&med, item.MedicineID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Medicine not found"})
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
	}

	order := models.Order{
		UserID:     userID,
		Address:    address,
		Status:     "confirmed",
		TotalPrice: total,
		Items:      orderItems,
	}

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to place order"})
		return
	}

	fmt.Println("✅ Order placed! ID:", order.ID)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Payment successful! Order placed.",
		"payment_id": paymentID,
		"order_id":   order.ID,
		"total":      total,
	})
}

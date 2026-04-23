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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact Razorpay"})
		return
	}
	defer resp.Body.Close()

	var razorpayOrder map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&razorpayOrder)

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

	if expectedSignature != input.RazorpaySignature {
		userID := c.GetUint("user_id")
		config.DB.Create(&models.Payment{
			UserID:            userID,
			RazorpayOrderID:   input.RazorpayOrderID,
			RazorpayPaymentID: input.RazorpayPaymentID,
			Status:            "failed",
			Method:            "razorpay",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment verification failed"})
		return
	}

	PlaceOrderAfterPayment(c, input.Address, input.Items, input.RazorpayOrderID, input.RazorpayPaymentID)
}

// fetchRazorpayMethod calls Razorpay's API to get the actual payment method (card, upi, netbanking, wallet)
func fetchRazorpayMethod(paymentID string) string {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.razorpay.com/v1/payments/%s", paymentID),
		nil,
	)
	if err != nil {
		fmt.Println("⚠️ Could not build Razorpay fetch request:", err)
		return "razorpay"
	}

	req.SetBasicAuth(razorpayKeyID, razorpayKeySecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("⚠️ Could not reach Razorpay to fetch method:", err)
		return "razorpay"
	}
	defer resp.Body.Close()

	var payDetails map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payDetails); err != nil {
		fmt.Println("⚠️ Could not decode Razorpay payment details:", err)
		return "razorpay"
	}

	if method, ok := payDetails["method"].(string); ok && method != "" {
		fmt.Println("💳 Razorpay payment method:", method)
		return method // "card", "upi", "netbanking", "wallet", "emi"
	}

	return "razorpay"
}

func PlaceOrderAfterPayment(
	c *gin.Context,
	address string,
	items []struct {
		MedicineID uint `json:"medicine_id"`
		Quantity   int  `json:"quantity"`
	},
	razorpayOrderID string,
	paymentID string,
) {
	userID := c.GetUint("user_id")

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

	// ✅ Fetch real payment method from Razorpay (card / upi / netbanking / wallet)
	method := fetchRazorpayMethod(paymentID)

	// ✅ Save successful payment record with correct method
	config.DB.Create(&models.Payment{
		UserID:            userID,
		OrderID:           order.ID,
		RazorpayOrderID:   razorpayOrderID,
		RazorpayPaymentID: paymentID,
		Amount:            total,
		Status:            "success",
		Method:            method,
	})

	fmt.Println("✅ Order placed! ID:", order.ID, "| Method:", method)
	c.JSON(http.StatusOK, gin.H{
		"message":    "Payment successful! Order placed.",
		"payment_id": paymentID,
		"order_id":   order.ID,
		"total":      total,
		"method":     method,
	})
}

// ── Get Payment History ────────────────────────────────────────────────────
func GetPaymentHistory(c *gin.Context) {
	userID := c.GetUint("user_id")
	fmt.Println("💳 Fetching payment history for user:", userID)

	var payments []models.Payment
	result := config.DB.
		Preload("Order.Items.Medicine").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&payments)

	fmt.Println("💳 Payments found:", result.RowsAffected)
	c.JSON(http.StatusOK, payments)
}

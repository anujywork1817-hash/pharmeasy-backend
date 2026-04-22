package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// ── In-memory OTP store ────────────────────────────────────────────────────
type otpEntry struct {
	OTP       string
	ExpiresAt time.Time
}

var (
	otpStore = make(map[string]otpEntry)
	otpMu    sync.Mutex
)

func Register(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Phone    string `json:"phone" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Address  string `json:"address"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Name:     input.Name,
		Email:    input.Email,
		Phone:    input.Phone,
		Password: string(hash),
		Address:  input.Address,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Phone number already registered"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "pharmeasy_super_secret_key_2024"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	fmt.Println("✅ Token generated for user:", user.Email)

	c.JSON(http.StatusOK, gin.H{
		"token": tokenStr,
		"user":  user,
	})
}

func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ── Send OTP ───────────────────────────────────────────────────────────────
func SendForgotPasswordOTP(c *gin.Context) {
	var input struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", input.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "No account found with this phone number"})
		return
	}

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	otpMu.Lock()
	otpStore[input.Phone] = otpEntry{
		OTP:       otp,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	otpMu.Unlock()

	fmt.Println("📱 OTP for", input.Phone, ":", otp)

	// Try SMS but don't fail if it doesn't work
	if err := sendSMSFast2SMS(input.Phone, otp); err != nil {
		fmt.Println("❌ SMS error (non-fatal):", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP sent successfully",
		"otp":     otp, // ⚠️ remove in production
	})
}

// ── Verify OTP + Reset Password ────────────────────────────────────────────
func VerifyOTPAndResetPassword(c *gin.Context) {
	var input struct {
		Phone       string `json:"phone" binding:"required"`
		OTP         string `json:"otp" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	otpMu.Lock()
	entry, exists := otpStore[input.Phone]
	otpMu.Unlock()

	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP not found. Request a new one."})
		return
	}

	if time.Now().After(entry.ExpiresAt) {
		otpMu.Lock()
		delete(otpStore, input.Phone)
		otpMu.Unlock()
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP expired. Request a new one."})
		return
	}

	if entry.OTP != input.OTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	if err := config.DB.Model(&models.User{}).
		Where("phone = ?", input.Phone).
		Update("password", string(hash)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	otpMu.Lock()
	delete(otpStore, input.Phone)
	otpMu.Unlock()

	fmt.Println("✅ Password reset for phone:", input.Phone)
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// ── Fast2SMS helper ────────────────────────────────────────────────────────
func sendSMSFast2SMS(phone, otp string) error {
	apiKey := "z2FjgotPbw4yr1cxdlNEi9SaCOmGW7DknXRu8eYZHAKMTQ6UsIVGizHjuOKFEewJ4dkxS85aBQXM0ZWA"

	payload := map[string]string{
		"route":            "otp",
		"variables_values": otp,
		"numbers":          phone,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST",
		"https://www.fast2sms.com/dev/bulkV2",
		bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	fmt.Println("📨 Fast2SMS response:", result)

	if resp.StatusCode != 200 {
		return fmt.Errorf("fast2sms error: %v", result)
	}

	return nil
}

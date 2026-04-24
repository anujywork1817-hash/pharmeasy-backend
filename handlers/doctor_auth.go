package handlers

import (
	"fmt"
	"net/http"
	"os"
	"pharmeasy-backend/config"
	"pharmeasy-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func DoctorRegister(c *gin.Context) {
	var input struct {
		Name      string `json:"name" binding:"required"`
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required,min=6"`
		Phone     string `json:"phone" binding:"required"`
		Specialty string `json:"specialty" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Doctor
	if err := config.DB.Where("email = ?", input.Email).
		First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	doctor := models.Doctor{
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hash),
		Phone:     input.Phone,
		Specialty: input.Specialty,
	}

	if err := config.DB.Create(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create doctor"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Doctor registered successfully"})
}

func DoctorLogin(c *gin.Context) {
	var input struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var doctor models.Doctor
	if err := config.DB.Where("email = ?", input.Email).
		First(&doctor).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(doctor.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "pharmeasy_super_secret_key_2024"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"doctor_id": doctor.ID,
		"email":     doctor.Email,
		"role":      "doctor",
		"exp":       time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	fmt.Println("✅ Doctor login:", doctor.Email)

	c.JSON(http.StatusOK, gin.H{
		"token":  tokenStr,
		"doctor": doctor,
	})
}

func DoctorGetProfile(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")
	var doctor models.Doctor
	if err := config.DB.First(&doctor, doctorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}
	c.JSON(http.StatusOK, doctor)
}

func DoctorUpdateProfile(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")

	var input struct {
		About       string `json:"about"`
		Fee         int    `json:"fee"`
		Location    string `json:"location"`
		UpiID       string `json:"upi_id"`
		BankAccount string `json:"bank_account"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&models.Doctor{}).Where("id = ?", doctorID).Updates(input)
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"pharmeasy-backend/config"
	"pharmeasy-backend/models"
	"strconv"
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

// ✅ FIXED: Now accepts multipart/form-data so profile image can be uploaded
func DoctorUpdateProfile(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")

	// Parse multipart form (max 10 MB)
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		// Not multipart — fall back to JSON
		doctorUpdateProfileJSON(c, doctorID)
		return
	}

	updates := map[string]interface{}{}

	// ── Text fields ──────────────────────────────────────────────────────────
	if v := c.PostForm("name"); v != "" {
		updates["name"] = v
	}
	if v := c.PostForm("phone"); v != "" {
		updates["phone"] = v
	}
	if v := c.PostForm("specialty"); v != "" {
		updates["specialty"] = v
	}
	if v := c.PostForm("about"); v != "" {
		updates["about"] = v
	}
	if v := c.PostForm("location"); v != "" {
		updates["location"] = v
	}
	if v := c.PostForm("upi_id"); v != "" {
		updates["upi_id"] = v
	}
	if v := c.PostForm("bank_account"); v != "" {
		updates["bank_account"] = v
	}
	if v := c.PostForm("experience"); v != "" {
		if exp, err := strconv.Atoi(v); err == nil && exp > 0 {
			updates["experience"] = exp
		}
	}
	if v := c.PostForm("fee"); v != "" {
		if fee, err := strconv.Atoi(v); err == nil && fee > 0 {
			updates["fee"] = fee
		}
	}
	if v := c.PostForm("is_available_today"); v != "" {
		updates["is_available_today"] = v == "true"
	}

	// ── Profile image ─────────────────────────────────────────────────────────
	file, header, err := c.Request.FormFile("profile_image")
	if err == nil {
		defer file.Close()

		// Make sure uploads dir exists
		uploadDir := "./uploads/doctors"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		// Unique filename: doctorID_timestamp.ext
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("%d_%d%s", doctorID, time.Now().Unix(), ext)
		savePath := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(header, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}

		// Store public URL path
		updates["profile_image"] = fmt.Sprintf("/uploads/doctors/%s", filename)
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	if err := config.DB.Model(&models.Doctor{}).
		Where("id = ?", doctorID).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	var doctor models.Doctor
	if err := config.DB.First(&doctor, doctorID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated profile"})
		return
	}

	fmt.Println("✅ Doctor profile updated:", doctor.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"doctor":  doctor,
	})
}

// Fallback: original JSON-only update (no image)
func doctorUpdateProfileJSON(c *gin.Context, doctorID uint) {
	var input struct {
		Name             string `json:"name"`
		Phone            string `json:"phone"`
		Specialty        string `json:"specialty"`
		Experience       int    `json:"experience"`
		About            string `json:"about"`
		Fee              int    `json:"fee"`
		Location         string `json:"location"`
		UpiID            string `json:"upi_id"`
		BankAccount      string `json:"bank_account"`
		IsAvailableToday bool   `json:"is_available_today"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"is_available_today": input.IsAvailableToday,
	}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Phone != "" {
		updates["phone"] = input.Phone
	}
	if input.Specialty != "" {
		updates["specialty"] = input.Specialty
	}
	if input.Experience > 0 {
		updates["experience"] = input.Experience
	}
	if input.About != "" {
		updates["about"] = input.About
	}
	if input.Fee > 0 {
		updates["fee"] = input.Fee
	}
	if input.Location != "" {
		updates["location"] = input.Location
	}
	if input.UpiID != "" {
		updates["upi_id"] = input.UpiID
	}
	if input.BankAccount != "" {
		updates["bank_account"] = input.BankAccount
	}

	if err := config.DB.Model(&models.Doctor{}).
		Where("id = ?", doctorID).
		Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	var doctor models.Doctor
	if err := config.DB.First(&doctor, doctorID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated profile"})
		return
	}

	fmt.Println("✅ Doctor profile updated:", doctor.Email)

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"doctor":  doctor,
	})
}

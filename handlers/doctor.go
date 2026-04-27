package handlers

import (
	"fmt"
	"net/http"
	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

// Get all doctors with filters
func GetDoctors(c *gin.Context) {
	var doctors []models.Doctor

	query := config.DB
	if specialty := c.Query("specialty"); specialty != "" && specialty != "All" {
		query = query.Where("specialty = ?", specialty)
	}
	if search := c.Query("search"); search != "" {
		query = query.Where("name ILIKE ? OR specialty ILIKE ? OR location ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if c.Query("available_today") == "true" {
		query = query.Where("is_available_today = ?", true)
	}
	if c.Query("online_only") == "true" {
		query = query.Where("is_available_online = ?", true)
	}
	if maxFee := c.Query("max_fee"); maxFee != "" {
		query = query.Where("fee <= ?", maxFee)
	}

	query.Find(&doctors)
	c.JSON(http.StatusOK, doctors)
}

// Get single doctor
func GetDoctorByID(c *gin.Context) {
	var doctor models.Doctor
	if err := config.DB.First(&doctor, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}
	c.JSON(http.StatusOK, doctor)
}

// Book appointment
func BookAppointment(c *gin.Context) {
	userID := c.GetUint("user_id")

	var input struct {
		DoctorID     uint   `json:"doctor_id" binding:"required"`
		Date         string `json:"date" binding:"required"`
		TimeSlot     string `json:"time_slot" binding:"required"`
		ConsultType  string `json:"consult_type" binding:"required"`
		PatientName  string `json:"patient_name" binding:"required"`
		PatientPhone string `json:"patient_phone" binding:"required"`
		Notes        string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if slot already booked
	var existing models.Appointment
	if err := config.DB.Where(
		"doctor_id = ? AND date = ? AND time_slot = ? AND status != ?",
		input.DoctorID, input.Date, input.TimeSlot, "cancelled",
	).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": fmt.Sprintf("This slot (%s on %s) is already booked. Please choose another time.", input.TimeSlot, input.Date),
		})
		return
	}

	appointment := models.Appointment{
		UserID:       userID,
		DoctorID:     input.DoctorID,
		Date:         input.Date,
		TimeSlot:     input.TimeSlot,
		ConsultType:  input.ConsultType,
		Status:       "confirmed",
		PatientName:  input.PatientName,
		PatientPhone: input.PatientPhone,
		Notes:        input.Notes,
	}

	if err := config.DB.Create(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to book appointment"})
		return
	}

	// Load doctor details
	config.DB.Preload("Doctor").First(&appointment, appointment.ID)

	fmt.Println("✅ Appointment booked:", appointment.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Appointment booked successfully!",
		"appointment": appointment,
	})
}

// Get user's appointments
func GetMyAppointments(c *gin.Context) {
	userID := c.GetUint("user_id")
	var appointments []models.Appointment
	config.DB.Preload("Doctor").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&appointments)
	c.JSON(http.StatusOK, appointments)
}

// Cancel appointment
func CancelAppointment(c *gin.Context) {
	userID := c.GetUint("user_id")
	appointmentID := c.Param("id")

	var appointment models.Appointment
	if err := config.DB.Where("id = ? AND user_id = ?", appointmentID, userID).
		First(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	config.DB.Model(&appointment).Update("status", "cancelled")
	c.JSON(http.StatusOK, gin.H{"message": "Appointment cancelled successfully"})
}

// Get booked slots for a doctor on a date
func GetBookedSlots(c *gin.Context) {
	doctorID := c.Param("id")
	date := c.Query("date")

	var appointments []models.Appointment
	config.DB.Where("doctor_id = ? AND date = ? AND status != ?",
		doctorID, date, "cancelled").
		Find(&appointments)

	bookedSlots := make([]string, 0)
	for _, a := range appointments {
		bookedSlots = append(bookedSlots, a.TimeSlot)
	}

	c.JSON(http.StatusOK, gin.H{"booked_slots": bookedSlots})
}

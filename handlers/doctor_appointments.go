package handlers

import (
	"net/http"
	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

// Get today's appointments for doctor
func DoctorGetTodayAppointments(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")

	var appointments []models.Appointment
	today := c.Query("date")
	if today == "" {
		today = "2026-04-24" // use actual today
	}

	config.DB.Preload("User").
		Where("doctor_id = ? AND date = ?", doctorID, today).
		Order("time_slot asc").
		Find(&appointments)

	c.JSON(http.StatusOK, appointments)
}

// Get all appointments for doctor
func DoctorGetAllAppointments(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")
	var appointments []models.Appointment
	config.DB.Preload("User").
		Where("doctor_id = ?", doctorID).
		Order("date desc, time_slot asc").
		Find(&appointments)
	c.JSON(http.StatusOK, appointments)
}

// Update appointment status
func DoctorUpdateAppointment(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")
	appointmentID := c.Param("id")

	var input struct {
		Status       string `json:"status"`
		DoctorNotes  string `json:"doctor_notes"`
		Prescription string `json:"prescription"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var appointment models.Appointment
	if err := config.DB.Where("id = ? AND doctor_id = ?",
		appointmentID, doctorID).First(&appointment).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	config.DB.Model(&appointment).Updates(map[string]interface{}{
		"status":       input.Status,
		"doctor_notes": input.DoctorNotes,
		"prescription": input.Prescription,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Appointment updated"})
}

// Get doctor earnings
func DoctorGetEarnings(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")

	var appointments []models.Appointment
	config.DB.Preload("Doctor").
		Where("doctor_id = ? AND status = ?", doctorID, "completed").
		Find(&appointments)

	var total float64
	monthly := make(map[string]float64)

	for _, a := range appointments {
		var doctor models.Doctor
		config.DB.First(&doctor, a.DoctorID)
		total += float64(doctor.Fee)

		month := a.Date[:7] // YYYY-MM
		monthly[month] += float64(doctor.Fee)
	}

	c.JSON(http.StatusOK, gin.H{
		"total_earnings":     total,
		"monthly_earnings":   monthly,
		"total_appointments": len(appointments),
	})
}

// Update doctor availability
func DoctorUpdateAvailability(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")

	var input struct {
		IsAvailableToday  bool     `json:"is_available_today"`
		IsAvailableOnline bool     `json:"is_available_online"`
		AvailableSlots    []string `json:"available_slots"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&models.Doctor{}).Where("id = ?", doctorID).Updates(map[string]interface{}{
		"is_available_today":  input.IsAvailableToday,
		"is_available_online": input.IsAvailableOnline,
		"available_slots":     input.AvailableSlots,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Availability updated"})
}

// Get doctor reviews
func DoctorGetReviews(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")
	var reviews []models.Review
	config.DB.Preload("User").
		Where("medicine_id = ?", doctorID).
		Order("created_at desc").
		Find(&reviews)
	c.JSON(http.StatusOK, reviews)
}

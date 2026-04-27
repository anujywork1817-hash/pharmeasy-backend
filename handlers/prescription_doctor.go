package handlers

import (
	"net/http"
	"pharmeasy-backend/config"
	"pharmeasy-backend/models"

	"github.com/gin-gonic/gin"
)

// POST /api/doctor/prescriptions
func DoctorSavePrescription(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")

	var input struct {
		PatientName   string `json:"patient_name" binding:"required"`
		PatientAge    string `json:"patient_age"`
		PatientGender string `json:"patient_gender"`
		PatientPhone  string `json:"patient_phone"`
		Diagnosis     string `json:"diagnosis" binding:"required"`
		Complaints    string `json:"complaints"`
		Medicines     string `json:"medicines" binding:"required"`
		Advice        string `json:"advice"`
		FollowUp      string `json:"follow_up"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prescription := models.DoctorPrescription{
		DoctorID:      doctorID,
		PatientName:   input.PatientName,
		PatientAge:    input.PatientAge,
		PatientGender: input.PatientGender,
		PatientPhone:  input.PatientPhone,
		Diagnosis:     input.Diagnosis,
		Complaints:    input.Complaints,
		Medicines:     input.Medicines,
		Advice:        input.Advice,
		FollowUp:      input.FollowUp,
	}

	if err := config.DB.Create(&prescription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save prescription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Prescription saved successfully",
		"prescription": prescription,
	})
}

// GET /api/doctor/prescriptions
func DoctorGetPrescriptions(c *gin.Context) {
	doctorID := c.GetUint("doctor_id")

	var prescriptions []models.DoctorPrescription
	if err := config.DB.
		Where("doctor_id = ?", doctorID).
		Order("created_at desc").
		Find(&prescriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch prescriptions"})
		return
	}

	c.JSON(http.StatusOK, prescriptions)
}

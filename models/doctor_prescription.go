package models

import "gorm.io/gorm"

type DoctorPrescription struct {
	gorm.Model
	DoctorID      uint   `json:"doctor_id"`
	PatientName   string `json:"patient_name"`
	PatientAge    string `json:"patient_age"`
	PatientGender string `json:"patient_gender"`
	PatientPhone  string `json:"patient_phone"`
	Diagnosis     string `json:"diagnosis"`
	Complaints    string `json:"complaints"`
	Medicines     string `json:"medicines"` // JSON string
	Advice        string `json:"advice"`
	FollowUp      string `json:"follow_up"`
}

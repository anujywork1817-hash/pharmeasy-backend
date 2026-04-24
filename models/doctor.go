package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Doctor struct {
	gorm.Model
	Name              string         `json:"name"`
	Email             string         `json:"email" gorm:"uniqueIndex"`
	Password          string         `json:"-"`
	Phone             string         `json:"phone"`
	Specialty         string         `json:"specialty"`
	Location          string         `json:"location"`
	Rating            float64        `json:"rating"`
	ReviewCount       int            `json:"review_count"`
	Experience        int            `json:"experience"`
	Fee               int            `json:"fee"`
	IsAvailableToday  bool           `json:"is_available_today"`
	IsAvailableOnline bool           `json:"is_available_online"`
	ImageURL          string         `json:"image_url"`
	AvailableSlots    pq.StringArray `json:"available_slots" gorm:"type:text[]"`
	About             string         `json:"about"`
	Qualifications    pq.StringArray `json:"qualifications" gorm:"type:text[]"`
	IsVerified        bool           `json:"is_verified"`
	BankAccount       string         `json:"bank_account"`
	UpiID             string         `json:"upi_id"`
}

type Appointment struct {
	gorm.Model
	UserID       uint   `json:"user_id"`
	User         User   `json:"user"`
	DoctorID     uint   `json:"doctor_id"`
	Doctor       Doctor `json:"doctor"`
	Date         string `json:"date"`
	TimeSlot     string `json:"time_slot"`
	ConsultType  string `json:"consult_type"`
	Status       string `json:"status"`
	PatientName  string `json:"patient_name"`
	PatientPhone string `json:"patient_phone"`
	Notes        string `json:"notes"`
	DoctorNotes  string `json:"doctor_notes"`
	Prescription string `json:"prescription"`
}

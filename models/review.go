package models

import "gorm.io/gorm"

type Review struct {
	gorm.Model
	UserID     uint    `json:"user_id" gorm:"column:user_id"`
	User       User    `json:"user"`
	MedicineID uint    `json:"medicine_id" gorm:"column:medicine_id"`
	Rating     float64 `json:"rating" gorm:"column:rating"`
	Comment    string  `json:"comment" gorm:"column:comment"`
}

func (Review) TableName() string {
	return "reviews"
}

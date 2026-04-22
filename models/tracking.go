package models

import "gorm.io/gorm"

type TrackingUpdate struct {
	gorm.Model
	OrderID     uint   `json:"order_id" gorm:"column:order_id"`
	Status      string `json:"status" gorm:"column:status"`
	Description string `json:"description" gorm:"column:description"`
	Location    string `json:"location" gorm:"column:location"`
}

func (TrackingUpdate) TableName() string {
	return "tracking_updates"
}

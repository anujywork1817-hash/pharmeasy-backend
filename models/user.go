package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name"`
	Email    string `json:"email" gorm:"unique"`
	Phone    string `json:"phone" gorm:"unique"`
	Password string `json:"-"`
	Address  string `json:"address"`
}

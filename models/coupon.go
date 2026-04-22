package models

import "gorm.io/gorm"

type Coupon struct {
	gorm.Model
	Code           string  `json:"code" gorm:"column:code;unique"`
	Description    string  `json:"description" gorm:"column:description"`
	DiscountType   string  `json:"discount_type" gorm:"column:discount_type"` // 'percent' or 'flat'
	DiscountValue  float64 `json:"discount_value" gorm:"column:discount_value"`
	MinOrderAmount float64 `json:"min_order_amount" gorm:"column:min_order_amount"`
	MaxDiscount    float64 `json:"max_discount" gorm:"column:max_discount"`
	UsageLimit     int     `json:"usage_limit" gorm:"column:usage_limit"`
	UsedCount      int     `json:"used_count" gorm:"column:used_count"`
	IsActive       bool    `json:"is_active" gorm:"column:is_active"`
	ExpiresAt      string  `json:"expires_at" gorm:"column:expires_at"`
}

func (Coupon) TableName() string {
	return "coupons"
}

type CouponUsage struct {
	gorm.Model
	CouponID uint    `json:"coupon_id" gorm:"column:coupon_id"`
	Coupon   Coupon  `json:"coupon"`
	UserID   uint    `json:"user_id" gorm:"column:user_id"`
	OrderID  uint    `json:"order_id" gorm:"column:order_id"`
	Discount float64 `json:"discount" gorm:"column:discount"`
}

func (CouponUsage) TableName() string {
	return "coupon_usages"
}

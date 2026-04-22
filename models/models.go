package models

import "gorm.io/gorm"

type Payment struct {
    gorm.Model
    UserID            uint    `json:"user_id" gorm:"column:user_id"`
    OrderID           uint    `json:"order_id" gorm:"column:order_id"`
    Order             Order   `json:"order"`
    RazorpayOrderID   string  `json:"razorpay_order_id" gorm:"column:razorpay_order_id"`
    RazorpayPaymentID string  `json:"razorpay_payment_id" gorm:"column:razorpay_payment_id"`
    Amount            float64 `json:"amount" gorm:"column:amount"`
    Status            string  `json:"status" gorm:"column:status"`
    Method            string  `json:"method" gorm:"column:method"`
}

func (Payment) TableName() string {
    return "payments"
}

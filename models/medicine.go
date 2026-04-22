package models

import "gorm.io/gorm"

type Medicine struct {
	gorm.Model
	Name        string  `json:"name" gorm:"column:name"`
	Brand       string  `json:"brand" gorm:"column:brand"`
	Description string  `json:"description" gorm:"column:description"`
	Price       float64 `json:"price" gorm:"column:price"`
	Discount    float64 `json:"discount" gorm:"column:discount"`
	Stock       int     `json:"stock" gorm:"column:stock"`
	RequiresRx  bool    `json:"requires_rx" gorm:"column:requires_rx"`
	ImageURL    string  `json:"image_url" gorm:"column:image_url"`
	Category    string  `json:"category" gorm:"column:category"`
}

func (Medicine) TableName() string {
	return "medicines"
}

type Order struct {
	gorm.Model
	UserID          uint             `json:"user_id" gorm:"column:user_id"`
	User            User             `json:"user"`
	Status          string           `json:"status" gorm:"column:status"`
	TotalPrice      float64          `json:"total_price" gorm:"column:total_price"`
	DiscountAmount  float64          `json:"discount_amount" gorm:"column:discount_amount"`
	FinalPrice      float64          `json:"final_price" gorm:"column:final_price"`
	Address         string           `json:"address" gorm:"column:address"`
	PaymentID       string           `json:"payment_id" gorm:"column:payment_id"`
	PaymentMethod   string           `json:"payment_method" gorm:"column:payment_method"`
	CouponCode      string           `json:"coupon_code" gorm:"column:coupon_code"`
	Items           []OrderItem      `json:"items"`
	TrackingUpdates []TrackingUpdate `json:"tracking_updates"`
}

func (Order) TableName() string {
	return "orders"
}

type OrderItem struct {
	gorm.Model
	OrderID    uint     `json:"order_id" gorm:"column:order_id"`
	MedicineID uint     `json:"medicine_id" gorm:"column:medicine_id"`
	Medicine   Medicine `json:"Medicine"`
	Quantity   int      `json:"quantity" gorm:"column:quantity"`
	Price      float64  `json:"price" gorm:"column:price"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

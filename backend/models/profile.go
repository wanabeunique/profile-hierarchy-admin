package models

import "time"

type Profile struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Type            string     `gorm:"not null;index;index:idx_type_parent" json:"type"`
	ParentID        *uint      `gorm:"index;index:idx_type_parent" json:"parentId"`
	Name            string     `gorm:"not null" json:"name"`
	Email           string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash    string     `gorm:"not null" json:"-"`
	Commission      float64    `gorm:"default:0" json:"commission"`
	PaymentDeadline *time.Time `json:"paymentDeadline"`
	CreatedAt       time.Time  `json:"createdAt"`

	// Relations
	Parent   *Profile  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Profile `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Orders   []Order   `gorm:"foreignKey:ProfileID" json:"orders,omitempty"`

	// Computed fields (not stored in DB)
	TotalAmount     float64 `gorm:"-" json:"totalAmount"`
	TotalPaid       float64 `gorm:"-" json:"totalPaid"`
	TotalRemaining  float64 `gorm:"-" json:"totalRemaining"`
	OrderCount      int     `gorm:"-" json:"orderCount"`
	SubAccountCount int     `gorm:"-" json:"subAccountCount"`
}

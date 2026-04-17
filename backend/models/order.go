package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProfileID uint      `gorm:"index;not null" json:"profileId"`
	Number    string    `gorm:"not null" json:"number"`
	Amount    float64   `gorm:"not null" json:"amount"`
	Paid      float64   `gorm:"default:0" json:"paid"`
	CreatedAt time.Time `json:"createdAt"`

	// Computed fields (not stored in DB)
	Remaining float64 `gorm:"-" json:"remaining"`
	Status    string  `gorm:"-" json:"status"`
}

func (o *Order) AfterFind(tx *gorm.DB) error {
	o.ComputeFields()
	return nil
}

// ComputeFields sets the derived Remaining and Status values.
func (o *Order) ComputeFields() {
	o.Remaining = o.Amount - o.Paid
	if o.Remaining < 0 {
		o.Remaining = 0
	}

	switch {
	case o.Paid <= 0:
		o.Status = "Не оплачено"
	case o.Paid < o.Amount:
		o.Status = "Частично оплачено"
	default:
		o.Status = "Оплачено"
	}
}

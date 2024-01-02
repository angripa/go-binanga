package model

import (
	accountModel "binanga/internal/account/model"
	"time"
)

type Merchant struct {
	ID        uint      `gorm:"column:id"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	DeletedAt int64     `gorm:"column:deleted_at"`
	User      accountModel.Account
	UserID    uint
}

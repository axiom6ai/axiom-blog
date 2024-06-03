package model

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;column:ID" json:"ID"`
	UserName  string    `gorm:"column:username"`
	Email     string
	Passcode  string
	Passwd    string
	Nickname  string
	Avatar    string
	Gender    uint8
	Introduce string
	State     uint8
	IsRoot    int `gorm:"column:is_root"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

package model

import "time"

type URL struct {
	ID        uint   `gorm:"primaryKey"`
	ShortURL  string `gorm:"unique"`
	LongURL   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

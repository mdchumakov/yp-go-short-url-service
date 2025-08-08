package model

import "time"

type URLsModel struct {
	ID        uint      `json:"id" db:"id"`
	ShortURL  string    `json:"short_url" db:"short_url"`
	LongURL   string    `json:"long_url" db:"long_url"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

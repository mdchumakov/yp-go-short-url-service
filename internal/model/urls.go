package model

import "time"

// URLsModel представляет модель URL в системе.
// Содержит информацию о коротком и длинном URL, статусе удаления и временных метках.
type URLsModel struct {
	ID        uint      `json:"id" db:"id"`
	ShortURL  string    `json:"short_url" db:"short_url"`
	LongURL   string    `json:"long_url" db:"long_url"`
	IsDeleted bool      `json:"is_deleted" db:"is_deleted"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

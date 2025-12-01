package model

import "time"

// UserModel представляет модель пользователя в системе.
// Содержит информацию об идентификаторе, имени, пароле, статусе анонимности и временных метках.
type UserModel struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Password    string    `json:"-" db:"password"`
	IsAnonymous bool      `json:"is_anonymous" db:"is_anonymous"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

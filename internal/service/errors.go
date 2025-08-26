package service

import "errors"

var (
	ErrURLWasDeleted    = errors.New("url was deleted")
	ErrURLAlreadyExists = errors.New("url already exists")
)

func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrURLAlreadyExists)
}

func IsDeletedError(err error) bool {
	return errors.Is(err, ErrURLWasDeleted)
}

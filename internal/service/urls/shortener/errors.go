package shortener

import "errors"

var ErrURLAlreadyExists = errors.New("url already exists")

func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrURLAlreadyExists)
}

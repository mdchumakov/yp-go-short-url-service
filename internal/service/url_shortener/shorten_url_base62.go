package url_shortener

import (
	"crypto/md5"
	"math/big"
)

const shortURLSize int = 8
const base62Chars string = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const hashSize int = 8

func shortenURLBase62(longURL string) string {
	hash := md5.Sum([]byte(longURL))

	num := new(big.Int)

	num.SetBytes(hash[:hashSize])

	shortURL := toBase62(num)
	return shortURL[:shortURLSize]
}

func toBase62(num *big.Int) string {
	if num.Cmp(big.NewInt(0)) == 0 {
		return "0"
	}

	result := ""
	base := big.NewInt(62)

	for num.Cmp(big.NewInt(0)) > 0 {
		remainder := new(big.Int)
		num.DivMod(num, base, remainder)
		result = string(base62Chars[remainder.Int64()]) + result
	}

	return result
}

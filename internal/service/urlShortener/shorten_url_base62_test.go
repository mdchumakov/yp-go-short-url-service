package urlShortener

import (
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_shortenURLBase62(t *testing.T) {
	tests := []struct {
		name     string
		longURL  string
		expected string
	}{
		{
			name:     "Test with a sample URL",
			longURL:  "https://example.com/some/long/url",
			expected: "4ZyG5E7z",
		},
		{
			name:     "Test with yandex practicum url",
			longURL:  "https://practicum.yandex.ru/",
			expected: "1BYWBNb1",
		},
		{
			name:     "Test with empty string",
			longURL:  "",
			expected: "d41d8cd9", // MD5 hash of empty string
		},
		{
			name:     "Test with single character",
			longURL:  "a",
			expected: "0cc175b9", // MD5 hash of "a"
		},
		{
			name:     "Test with very long URL",
			longURL:  strings.Repeat("https://example.com/very/long/url/", 100),
			expected: "8f14e45f", // MD5 hash of the repeated string
		},
		{
			name:     "Test with special characters",
			longURL:  "https://example.com/path?param=value&another=param#fragment",
			expected: "JxCQw7BL", // MD5 hash of the URL with special chars
		},
		{
			name:     "Test with unicode characters",
			longURL:  "https://example.com/путь/с/русскими/символами",
			expected: "8CsWIleA", // MD5 hash of unicode URL
		},
		{
			name:     "Test with numbers only",
			longURL:  "1234567890",
			expected: "e807f1fc", // MD5 hash of "1234567890"
		},
		{
			name:     "Test with mixed case",
			longURL:  "HTTPS://EXAMPLE.COM/MIXED/CASE/URL",
			expected: "9srVm05m", // MD5 hash of uppercase URL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortenURLBase62(tt.longURL)

			// Проверяем длину результата
			assert.Len(t, result, shortURLSize, "Short URL should have correct length")

			// Проверяем, что результат содержит только допустимые символы
			for _, char := range result {
				assert.Contains(t, base62Chars, string(char), "Short URL should contain only base62 characters")
			}

			// Проверяем детерминированность (одинаковый URL дает одинаковый результат)
			result2 := shortenURLBase62(tt.longURL)
			assert.Equal(t, result, result2, "Function should be deterministic")
		})
	}
}

func Test_shortenURLBase62_Properties(t *testing.T) {
	t.Run("deterministic behavior", func(t *testing.T) {
		url := "https://example.com/test"
		result1 := shortenURLBase62(url)
		result2 := shortenURLBase62(url)
		result3 := shortenURLBase62(url)

		assert.Equal(t, result1, result2, "Results should be identical for same input")
		assert.Equal(t, result2, result3, "Results should be identical for same input")
	})

	t.Run("different inputs produce different outputs", func(t *testing.T) {
		url1 := "https://example.com/test1"
		url2 := "https://example.com/test2"

		result1 := shortenURLBase62(url1)
		result2 := shortenURLBase62(url2)

		assert.NotEqual(t, result1, result2, "Different inputs should produce different outputs")
	})

	t.Run("output length consistency", func(t *testing.T) {
		testURLs := []string{
			"",
			"a",
			"https://example.com",
			strings.Repeat("very long url ", 1000),
		}

		for _, url := range testURLs {
			result := shortenURLBase62(url)
			assert.Len(t, result, shortURLSize, "All results should have same length")
		}
	})

	t.Run("output character set", func(t *testing.T) {
		testURLs := []string{
			"https://example.com",
			"https://google.com",
			"https://github.com",
		}

		for _, url := range testURLs {
			result := shortenURLBase62(url)
			for _, char := range result {
				assert.Contains(t, base62Chars, string(char),
					"All characters should be from base62 charset")
			}
		}
	})
}

func Test_toBase62(t *testing.T) {
	tests := []struct {
		name     string
		input    *big.Int
		expected string
	}{
		{
			name:     "Test with zero",
			input:    big.NewInt(0),
			expected: "0",
		},
		{
			name:     "Test with small number",
			input:    big.NewInt(123),
			expected: "1z",
		},
		{
			name:     "Test with large number",
			input:    big.NewInt(123456789),
			expected: "8M0kX",
		},
		{
			name:     "Test with number 61",
			input:    big.NewInt(61),
			expected: "z",
		},
		{
			name:     "Test with number 62",
			input:    big.NewInt(62),
			expected: "10",
		},
		{
			name:     "Test with number 63",
			input:    big.NewInt(63),
			expected: "11",
		},
		{
			name:     "Test with number 3843", // 61 * 62 + 61 = zz
			input:    big.NewInt(3843),
			expected: "zz",
		},
		{
			name:     "Test with number 3844", // 62 * 62 = 100
			input:    big.NewInt(3844),
			expected: "100",
		},
		{
			name:     "Test with very large number",
			input:    big.NewInt(999999999),
			expected: "15ftgF",
		},
		{
			name:     "Test with negative number (should handle gracefully)",
			input:    big.NewInt(-123),
			expected: "", // toBase62 returns empty string for negative numbers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toBase62(tt.input)
			assert.Equal(t, tt.expected, result, "toBase62 should return expected result")

			// Проверяем, что результат содержит только допустимые символы
			for _, char := range result {
				assert.Contains(t, base62Chars, string(char),
					"Result should contain only base62 characters")
			}
		})
	}
}

func Test_toBase62_Properties(t *testing.T) {
	t.Run("zero handling", func(t *testing.T) {
		result := toBase62(big.NewInt(0))
		assert.Equal(t, "0", result, "Zero should be represented as '0'")
	})

	t.Run("single digit numbers", func(t *testing.T) {
		for i := int64(0); i < 62; i++ {
			result := toBase62(big.NewInt(i))
			assert.Len(t, result, 1, "Single digit numbers should have length 1")
			assert.Contains(t, base62Chars, result, "Result should be in base62 charset")
		}
	})

	t.Run("two digit numbers", func(t *testing.T) {
		// Test some two-digit numbers
		testCases := []int64{62, 63, 124, 125}
		for _, num := range testCases {
			result := toBase62(big.NewInt(num))
			assert.Len(t, result, 2, "Two digit numbers should have length 2")
		}
	})

	t.Run("negative numbers", func(t *testing.T) {
		// toBase62 should handle negative numbers gracefully
		result := toBase62(big.NewInt(-1))
		assert.Equal(t, "", result, "Negative numbers should return empty string")

		result = toBase62(big.NewInt(-100))
		assert.Equal(t, "", result, "Negative numbers should return empty string")
	})

	t.Run("large numbers", func(t *testing.T) {
		// Test with a very large number
		largeNum := new(big.Int)
		largeNum.SetString("123456789012345678901234567890", 10)

		result := toBase62(largeNum)
		assert.NotEmpty(t, result, "Large numbers should produce non-empty result")

		// Check that all characters are valid
		for _, char := range result {
			assert.Contains(t, base62Chars, string(char),
				"All characters should be from base62 charset")
		}
	})
}

func Test_toBase62_EdgeCases(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		// This should panic with nil input
		assert.Panics(t, func() {
			toBase62(nil)
		}, "Function should panic with nil input")
	})

	t.Run("maximum uint64", func(t *testing.T) {
		maxUint64 := new(big.Int).SetUint64(18446744073709551615)
		result := toBase62(maxUint64)

		assert.NotEmpty(t, result, "Maximum uint64 should produce non-empty result")
		assert.Len(t, result, 11, "Maximum uint64 should produce 11-character result")
	})

	t.Run("powers of 62", func(t *testing.T) {
		// Test powers of 62 - create new big.Int for each test
		expectedResults := []string{"1", "10", "100", "1000", "10000"}
		values := []int64{1, 62, 3844, 238328, 14776336}

		for i, value := range values {
			result := toBase62(big.NewInt(value))
			assert.Equal(t, expectedResults[i], result,
				"Power of 62 should be represented correctly")
		}
	})
}

package utils

import "os"

// CheckFileExists проверяет существование файла по указанному пути.
// Возвращает true, если файл существует, и false в противном случае.
func CheckFileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

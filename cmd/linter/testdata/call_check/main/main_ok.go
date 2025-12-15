package main

import (
	"log"
	"os"
)

func main() {
	log.Fatal("error") // ok: вызов в функции main пакета main разрешен
	os.Exit(1)         // ok: вызов в функции main пакета main разрешен
}

func otherFunction() {
	log.Fatal("error") // want "использование log.Fatal запрещено вне функции main пакета main"
	os.Exit(1)         // want "использование os.Exit запрещено вне функции main пакета main"
}

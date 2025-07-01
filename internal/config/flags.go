package config

import "flag"

type Flags struct {
	ServerAddress string
	BaseURL       string
}

func NewFlags() *Flags {
	connectionAddr := flag.String(
		"a",
		"",
		"Отвечает за адрес запуска HTTP-сервера (значение может быть таким: localhost:8888)",
	)
	redirectURL := flag.String(
		"b",
		"",
		"Отвечает за базовый адрес результирующего сокращённого URL "+
			"(значение: адрес сервера перед коротким URL, например, http://localhost:8000/qsd54gFg).",
	)

	flag.Parse()

	return &Flags{
		ServerAddress: *connectionAddr,
		BaseURL:       *redirectURL,
	}
}

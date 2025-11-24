package config

import "flag"

type Flags struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	AuditFile       string
	AuditURL        string
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

	fileStoragePath := flag.String(
		"f",
		"",
		"Путь до файла, куда сохраняются данные в формате JSON. "+
			"Имя файла для значения по умолчанию придумайте сами.",
	)

	databaseDSN := flag.String(
		"d",
		"",
		"Строка с адресом подключения к БД",
	)

	auditFile := flag.String(
		"audit-file",
		"",
		"путь к файлу-приёмнику, в который сохраняются логи аудита",
	)
	auditURL := flag.String(
		"audit-url",
		"",
		"полный URL удаленного сервера-приёмника, куда отправляются логи аудита",
	)

	flag.Parse()

	return &Flags{
		ServerAddress:   *connectionAddr,
		BaseURL:         *redirectURL,
		FileStoragePath: *fileStoragePath,
		DatabaseDSN:     *databaseDSN,
		AuditFile:       *auditFile,
		AuditURL:        *auditURL,
	}
}

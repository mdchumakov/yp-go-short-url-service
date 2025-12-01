package config

const defaultAuditFilePath = ""
const defaultAuditURL = ""

// AuditSettings содержит настройки аудита приложения.
// Определяет пути для сохранения логов аудита в файл и URL для отправки на удаленный сервер.
type AuditSettings struct {
	File string `envconfig:"FILE" default:"" required:"false"`
	URL  string `envconfig:"URL" default:"" required:"false"`
}

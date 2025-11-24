package config

const defaultAuditFilePath = ""
const defaultAuditURL = ""

type AuditSettings struct {
	File string `envconfig:"FILE" default:"" required:"false"`
	URL  string `envconfig:"URL" default:"" required:"false"`
}

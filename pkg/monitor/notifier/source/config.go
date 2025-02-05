package source

import "errors"

type Config struct {
	SourceType         SourceType  `yaml:"type"`
	Name               string      `yaml:"name"`
	Group              *string     `yaml:"group,omitempty"`
	IncludeMonitorName bool        `yaml:"includeMonitorName"`
	Config             *RawMessage `yaml:"config"`
}

type SourceType string

const (
	SourceTypeUnknown  SourceType = "unknown"
	SourceTypeDiscord  SourceType = "discord"
	SourceTypeSMTP     SourceType = "smtp"
	SourceTypeSES      SourceType = "ses"
	SourceTypeTelegram SourceType = "telegram"
)

func (c *Config) Validate() error {
	if c.SourceType == SourceTypeUnknown {
		return errors.New("notifier source type is required")
	}

	return nil
}

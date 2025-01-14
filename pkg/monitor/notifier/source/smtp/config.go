package smtp

import (
	"errors"

	"github.com/creasty/defaults"
)

// Config represents the configuration for the email source
type Config struct {
	// SMTP server host
	Host string `yaml:"host" default:"localhost"`
	// SMTP server port
	Port int `yaml:"port" default:"587"`
	// Username for SMTP authentication
	Username string `yaml:"username"`
	// Password for SMTP authentication
	Password string `yaml:"password"`
	// From email address
	From string `yaml:"from"`
	// To email addresses
	To []string `yaml:"to"`
	// Whether to use TLS
	TLS bool `yaml:"tls" default:"false"`
	// Whether to skip TLS certificate verification
	InsecureSkipVerify bool `yaml:"insecureSkipVerify" default:"false"`
}

// SetDefaults sets the default values for the config
func (c *Config) SetDefaults() error {
	if err := defaults.Set(c); err != nil {
		return err
	}

	return nil
}

// Validate validates the config
func (c *Config) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}

	if c.From == "" {
		return errors.New("from is required")
	}

	if len(c.To) == 0 {
		return errors.New("to is required")
	}

	return nil
}

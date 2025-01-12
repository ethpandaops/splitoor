package controller

import "errors"

type Config struct {
	ControllerType ControllerType `yaml:"type"`
	Config         *RawMessage    `yaml:"config"`
}

type ControllerType string

const (
	ControllerTypeUnknown ControllerType = "unknown"
	ControllerTypeEOA     ControllerType = "eoa"
	ControllerTypeSafe    ControllerType = "safe"
)

func (c *Config) Validate() error {
	if c.ControllerType == ControllerTypeUnknown {
		return errors.New("controller type is required")
	}

	return nil
}

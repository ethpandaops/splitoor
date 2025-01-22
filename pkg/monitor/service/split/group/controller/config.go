package controller

import (
	"fmt"
)

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
	if c == nil {
		return fmt.Errorf("config is nil")
	}

	if c.ControllerType == ControllerTypeUnknown || c.ControllerType == "" {
		return fmt.Errorf("controller type is required")
	}

	return nil
}

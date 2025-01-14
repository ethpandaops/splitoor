package validator

import (
	"fmt"
	"time"
)

type Status struct {
	Timestamp time.Time
	Pubkey    string
	Status    string
	Source    string
	Group     string
	Monitor   string
}

const (
	StatusType = "validator_status"
)

func (v *Status) GetType() string {
	return StatusType
}

func (v *Status) GetGroup() string {
	return v.Group
}

func (v *Status) GetMonitor() string {
	return v.Monitor
}

func (v *Status) GetTitle() string {
	return fmt.Sprintf("[%s] %s validator has unexpectedly status", v.Monitor, v.Group)
}

func (v *Status) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Source: %s
Pubkey: %s
Status: %s`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.Source, v.Pubkey, v.Status)
}

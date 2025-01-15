package validator

import (
	"fmt"
	"time"
)

type Status struct {
	Timestamp time.Time
	Pubkey    string
	Status    string
	Group     string
	Monitor   string
}

const (
	StatusType = "validator_status"
)

func NewStatus(timestamp time.Time, status string, pubkey, group, monitor string) *Status {
	return &Status{
		Timestamp: timestamp,
		Pubkey:    pubkey,
		Status:    status,
		Group:     group,
		Monitor:   monitor,
	}
}

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
Pubkey: %s
Status: %s`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.Pubkey, v.Status)
}

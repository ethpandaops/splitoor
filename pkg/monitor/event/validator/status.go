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

func (v *Status) GetText() string {
	return fmt.Sprintf("Validator %s is %s", v.Pubkey, v.Status)
}

func (v *Status) GetMarkdown() string {
	return fmt.Sprintf("Validator %s is %s", v.Pubkey, v.Status)
}

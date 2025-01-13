package validator

import (
	"fmt"
	"time"
)

type MinBalance struct {
	Timestamp time.Time
	Pubkey    string
	Balance   uint64
	Source    string
	Group     string
}

const (
	MinBalanceType = "validator_min_balance"
)

func (v *MinBalance) GetType() string {
	return MinBalanceType
}

func (v *MinBalance) GetGroup() string {
	return v.Group
}

func (v *MinBalance) GetText() string {
	return fmt.Sprintf("Validator %s has a balance of %d", v.Pubkey, v.Balance)
}

func (v *MinBalance) GetMarkdown() string {
	return fmt.Sprintf("Validator %s has a balance of %d", v.Pubkey, v.Balance)
}

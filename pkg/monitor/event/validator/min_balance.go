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
	Monitor   string
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

func (v *MinBalance) GetMonitor() string {
	return v.Monitor
}

func (v *MinBalance) GetTitle() string {
	return fmt.Sprintf("[%s] %s validator has low balance", v.Monitor, v.Group)
}

func (v *MinBalance) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Source: %s
Pubkey: %s
Balance: %.4f ETH`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.Source, v.Pubkey, float64(v.Balance)/1e18)
}

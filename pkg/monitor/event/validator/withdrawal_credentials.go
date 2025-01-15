package validator

import (
	"fmt"
	"time"
)

type WithdrawalCredentials struct {
	Timestamp time.Time
	Pubkey    string
	Code      int64
	Group     string
	Monitor   string
}

const (
	WithdrawalCredentialsType = "validator_withdrawal_credentials"
)

func NewWithdrawalCredentials(timestamp time.Time, code int64, pubkey, group, monitor string) *WithdrawalCredentials {
	return &WithdrawalCredentials{
		Timestamp: timestamp,
		Pubkey:    pubkey,
		Code:      code,
		Group:     group,
		Monitor:   monitor,
	}
}

func (v *WithdrawalCredentials) GetType() string {
	return WithdrawalCredentialsType
}

func (v *WithdrawalCredentials) GetGroup() string {
	return v.Group
}

func (v *WithdrawalCredentials) GetMonitor() string {
	return v.Monitor
}

func (v *WithdrawalCredentials) GetTitle() string {
	return fmt.Sprintf("[%s] %s validator has unexpected withdrawal credentials type", v.Monitor, v.Group)
}

func (v *WithdrawalCredentials) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Pubkey: %s
Code: 0x%02x`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.Pubkey, v.Code)
}

package validator

import (
	"fmt"
	"time"
)

type TransactionQueue struct {
	Timestamp   time.Time
	SafeAddress string
	Source      string
	Group       string
	NumTxs      int
}

const (
	TransactionQueueType = "safe_transaction_queue"
)

func (v *TransactionQueue) GetType() string {
	return TransactionQueueType
}

func (v *TransactionQueue) GetGroup() string {
	return v.Group
}

func (v *TransactionQueue) GetText() string {
	return fmt.Sprintf("Safe %s has %d unexpected transaction(s) queued for execution", v.SafeAddress, v.NumTxs)
}

func (v *TransactionQueue) GetMarkdown() string {
	return fmt.Sprintf("Safe %s has %d unexpected transaction(s) queued for execution", v.SafeAddress, v.NumTxs)
}

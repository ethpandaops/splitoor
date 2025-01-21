package validator

import (
	"fmt"
	"time"
)

type Hash struct {
	Timestamp    time.Time
	SplitAddress string
	ExpectedHash string
	ActualHash   string
	Group        string
	Monitor      string
}

const (
	HashType = "split_hash"
)

func NewHash(timestamp time.Time, monitor, group, splitAddress, expectedHash, actualHash string) *Hash {
	return &Hash{
		Timestamp:    timestamp,
		SplitAddress: splitAddress,
		ExpectedHash: expectedHash,
		ActualHash:   actualHash,
		Group:        group,
		Monitor:      monitor,
	}
}

func (v *Hash) GetType() string {
	return HashType
}

func (v *Hash) GetGroup() string {
	return v.Group
}

func (v *Hash) GetMonitor() string {
	return v.Monitor
}

func (v *Hash) GetTitle() string {
	return fmt.Sprintf("[%s] %s split hash has changed", v.Monitor, v.Group)
}

func (v *Hash) GetDescription() string {
	return fmt.Sprintf(`
Timestamp: %s
Monitor: %s
Group: %s
Split Address: %s
Expected Hash: %s
Actual Hash: %s`, v.Timestamp.UTC().Format("2006-01-02 15:04:05 UTC"), v.Monitor, v.Group, v.SplitAddress, v.ExpectedHash, v.ActualHash)
}

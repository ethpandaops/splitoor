package validator

import (
	"fmt"
	"time"
)

type Hash struct {
	Timestamp    time.Time
	SplitAddress string
	SplitName    string
	Hash         string
	Source       string
	Group        string
}

const (
	HashType = "split_hash"
)

func (v *Hash) GetType() string {
	return HashType
}

func (v *Hash) GetGroup() string {
	return v.Group
}

func (v *Hash) GetText() string {
	return fmt.Sprintf("Split %s (%s) hash has changed to %s", v.SplitName, v.SplitAddress, v.Hash)
}

func (v *Hash) GetMarkdown() string {
	return fmt.Sprintf("Split %s (%s) hash has changed to %s", v.SplitName, v.SplitAddress, v.Hash)
}

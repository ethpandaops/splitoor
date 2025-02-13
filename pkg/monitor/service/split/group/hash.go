package group

import (
	"encoding/hex"

	spl "github.com/ethpandaops/splitoor/pkg/0xsplits/split"
)

func calculateHash(accounts []string, allocations []uint32) (string, error) {
	initialHashParams := &spl.HashParams{
		Accounts:              accounts,
		PercentageAllocations: allocations,
	}

	initialHash, err := spl.CalculateHash(initialHashParams)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(initialHash), nil
}

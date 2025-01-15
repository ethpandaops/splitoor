package split

import (
	"fmt"
)

func ParseRecipients(accounts []string, percentageAllocations []uint32) (recipients []string, allocations []uint32, err error) {
	if len(accounts) < 2 {
		return nil, nil, fmt.Errorf("must specify at least 2 recipients")
	}

	if len(accounts) != len(percentageAllocations) {
		return nil, nil, fmt.Errorf("number of accounts and percentage allocations must match")
	}

	allocations = make([]uint32, 0)

	var total uint32

	for _, v := range percentageAllocations {
		total += v
		allocations = append(allocations, v)
	}

	if total != 1000000 {
		return nil, nil, fmt.Errorf("percentage allocations must sum to 1000000, got %d", total)
	}

	return accounts, allocations, nil
}

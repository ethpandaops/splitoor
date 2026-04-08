package split

import (
	"errors"
	"fmt"
	"sort"
)

func ParseRecipients(accounts []string, percentageAllocations []uint32) (recipients []string, allocations []uint32, err error) {
	if len(accounts) < 2 {
		return nil, nil, errors.New("must specify at least 2 recipients")
	}

	if len(accounts) != len(percentageAllocations) {
		return nil, nil, errors.New("number of accounts and percentage allocations must match")
	}

	// Create pairs for sorting
	pairs := make([][2]any, len(accounts))
	for i := range accounts {
		pairs[i] = [2]any{accounts[i], percentageAllocations[i]}
	}

	// Sort by account address with safe type assertions
	sort.Slice(pairs, func(i, j int) bool {
		str1, ok1 := pairs[i][0].(string)
		str2, ok2 := pairs[j][0].(string)

		// If either conversion fails, we'll maintain stable sort order
		// This shouldn't happen with properly-formed data, but prevents panics
		if !ok1 || !ok2 {
			return i < j // Maintain stable ordering
		}

		return str1 < str2
	})

	// Separate back into sorted slices
	recipients = make([]string, len(accounts))
	allocations = make([]uint32, len(accounts))

	var total uint32

	for i, pair := range pairs {
		var ok bool

		recipients[i], ok = pair[0].(string)
		if !ok {
			return nil, nil, fmt.Errorf("invalid account address: %v", pair[0])
		}

		allocations[i], ok = pair[1].(uint32)
		if !ok {
			return nil, nil, fmt.Errorf("invalid allocation: %v", pair[1])
		}

		total += allocations[i]
	}

	if total != 1000000 {
		return nil, nil, fmt.Errorf("percentage allocations must sum to 1000000, got %d", total)
	}

	return recipients, allocations, nil
}

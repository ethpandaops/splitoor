package execution

import "math/big"

func averageBigInts(values []*big.Int) *big.Int {
	// Return 0 if values array is empty
	if len(values) == 0 {
		return big.NewInt(0)
	}

	sum := big.NewInt(0)
	validCount := 0

	// Add all non-nil values
	for _, v := range values {
		if v != nil {
			sum.Add(sum, v)

			validCount++
		}
	}

	// If no valid values were found, return 0
	if validCount == 0 {
		return big.NewInt(0)
	}

	// Safely divide by the number of valid values
	return sum.Div(sum, big.NewInt(int64(validCount)))
}

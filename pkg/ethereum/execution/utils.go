package execution

import "math/big"

func averageBigInts(values []*big.Int) *big.Int {
	if len(values) == 0 {
		return big.NewInt(0)
	}

	sum := big.NewInt(0)

	for _, v := range values {
		sum.Add(sum, v)
	}

	return sum.Div(sum, big.NewInt(int64(len(values))))
}

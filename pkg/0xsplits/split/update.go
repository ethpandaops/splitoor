package split

import (
	"context"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/crypto"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
)

type UpdateSplitParams struct {
	Accounts              []string
	PercentageAllocations []uint32
	DistributorFee        uint32

	mu sync.Mutex
}

func (p *UpdateSplitParams) order() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	accounts, allocations, err := ParseRecipients(p.Accounts, p.PercentageAllocations)
	if err != nil {
		return err
	}

	p.Accounts = accounts
	p.PercentageAllocations = allocations

	return nil
}

func (p *UpdateSplitParams) encode(splitAddress string) ([]interface{}, error) {
	// Create pairs for sorting
	pairs := make([][2]interface{}, len(p.Accounts))
	for i := range p.Accounts {
		pairs[i] = [2]interface{}{p.Accounts[i], p.PercentageAllocations[i]}
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
	accounts := make([]common.Address, len(p.Accounts))
	allocations := make([]uint32, len(p.PercentageAllocations))

	for i := range pairs {
		// Safely perform the type assertion with check
		accountStr, ok := pairs[i][0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid account type at index %d", i)
		}

		accounts[i] = common.HexToAddress(accountStr)

		allocation, ok := pairs[i][1].(uint32)
		if !ok {
			return nil, fmt.Errorf("invalid allocation type at index %d", i)
		}

		allocations[i] = allocation
	}

	return []interface{}{
		common.HexToAddress(splitAddress),
		accounts,
		allocations,
		p.DistributorFee,
	}, nil
}

func (c *Client) Update(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI, from, privateKey string, gasLimit uint64, params *UpdateSplitParams) (*string, error) {
	if err := params.order(); err != nil {
		return nil, err
	}

	if c.splitAddress == nil {
		return nil, fmt.Errorf("split address is not set")
	}

	splitAddress := *c.splitAddress // Dereference safely after nil check

	pKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	encodeParams, err := params.encode(splitAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to encode parameters: %w", err)
	}

	calldata, err := contractABI.EncodeMethodCalldata("updateSplit", encodeParams)
	if err != nil {
		return nil, err
	}

	txHash, err := node.WriteContract(ctx, c.contractAddress, calldata, from, pKey, big.NewInt(0), gasLimit)
	if err != nil {
		return nil, err
	}

	time.Sleep(time.Second * 5)

	return txHash, nil
}

package split

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/0xsequence/ethkit/go-ethereum/crypto"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
)

type DistributeETHParams struct {
	Accounts              []string
	PercentageAllocations []uint32
	DistributorFee        uint32
	DistributorAddress    string

	mu sync.Mutex
}

func (p *DistributeETHParams) order() error {
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

func (p *DistributeETHParams) encode(splitAddress string) ([]any, error) {
	// Create pairs for sorting
	pairs := make([][2]any, len(p.Accounts))
	for i := range p.Accounts {
		pairs[i] = [2]any{p.Accounts[i], p.PercentageAllocations[i]}
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

	distributorAddress := p.DistributorAddress
	if distributorAddress == "" {
		distributorAddress = "0x0000000000000000000000000000000000000000"
	}

	return []any{
		common.HexToAddress(splitAddress),
		accounts,
		allocations,
		p.DistributorFee,
		common.HexToAddress(distributorAddress),
	}, nil
}

func (c *Client) DistributeETH(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI, from, privateKey string, gasLimit uint64, params *DistributeETHParams) error {
	if err := params.order(); err != nil {
		return err
	}

	if c.splitAddress == nil {
		return errors.New("split address is not set")
	}

	splitAddress := *c.splitAddress // Dereference safely after nil check

	pKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return err
	}

	encodeParams, err := params.encode(splitAddress)
	if err != nil {
		return fmt.Errorf("failed to encode parameters: %w", err)
	}

	calldata, err := contractABI.EncodeMethodCalldata("distributeETH", encodeParams)
	if err != nil {
		return err
	}

	txHash, err := node.WriteContract(ctx, c.contractAddress, calldata, from, pKey, big.NewInt(0), gasLimit)
	if err != nil {
		return err
	}

	// Create a new context with timeout to prevent hanging indefinitely
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	c.log.WithField("tx", *txHash).Info("Waiting for transaction to be included in a block")

	if err = c.waitForTransaction(pollCtx, node, *txHash); err != nil {
		return err
	}

	receipt, err := node.TransactionReceipt(ctx, *txHash)
	if err != nil {
		return err
	}

	// Check for nil receipt
	if receipt == nil {
		return fmt.Errorf("received nil transaction receipt for tx hash %s", *txHash)
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return fmt.Errorf("transaction %s failed with status: %d", *txHash, receipt.Status)
	}

	return nil
}

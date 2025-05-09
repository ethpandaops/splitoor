package split

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/crypto"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
)

type CreateSplitParams struct {
	Controller            string
	Accounts              []string
	PercentageAllocations []uint32
	DistributorFee        uint32

	mu sync.Mutex
}

func (c *CreateSplitParams) order() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	accounts, allocations, err := ParseRecipients(c.Accounts, c.PercentageAllocations)
	if err != nil {
		return err
	}

	c.Accounts = accounts
	c.PercentageAllocations = allocations

	return nil
}

func (c *CreateSplitParams) encode() ([]interface{}, error) {
	// Create pairs for sorting
	pairs := make([][2]interface{}, len(c.Accounts))
	for i := range c.Accounts {
		pairs[i] = [2]interface{}{c.Accounts[i], c.PercentageAllocations[i]}
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
	accounts := make([]common.Address, len(c.Accounts))
	allocations := make([]uint32, len(c.PercentageAllocations))

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
		accounts,
		allocations,
		c.DistributorFee,
		common.HexToAddress(c.Controller),
	}, nil
}

func (c *Client) Create(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI, from, privateKey string, gasLimit uint64, params *CreateSplitParams) (*string, error) {
	if err := params.order(); err != nil {
		return nil, err
	}

	if c.splitAddress != nil {
		return nil, fmt.Errorf("split address is already set")
	}

	pKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	encodeParams, err := params.encode()
	if err != nil {
		return nil, fmt.Errorf("failed to encode parameters: %w", err)
	}

	calldata, err := contractABI.EncodeMethodCalldata("createSplit", encodeParams)
	if err != nil {
		return nil, err
	}

	txHash, err := node.WriteContract(ctx, c.contractAddress, calldata, from, pKey, big.NewInt(0), gasLimit)
	if err != nil {
		return nil, err
	}

	// Create a new context with timeout to prevent hanging indefinitely
	pollCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	c.log.WithField("tx", *txHash).Info("Waiting for transaction to be included in a block")

	// Add a ticker for controlled polling with backoff
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track attempts for exponential backoff
	attempt := 0
	maxAttempts := 600 // 10 minutes worth of 1-second attempts

	for attempt < maxAttempts {
		select {
		case <-pollCtx.Done():
			// Context cancelled or timed out
			return nil, fmt.Errorf("context cancelled or timed out while waiting for transaction: %w", pollCtx.Err())
		case <-ticker.C:
			// Time to check transaction status
			var isPending bool

			_, isPending, err = node.TransactionByHash(pollCtx, *txHash)
			if err != nil {
				// If we're getting transient errors, continue polling
				if strings.Contains(err.Error(), "not found") {
					attempt++

					continue
				}

				return nil, err
			}

			if !isPending {
				// Transaction is no longer pending, we can proceed
				goto TransactionComplete
			}

			// Increment attempt counter
			attempt++

			// Apply exponential backoff after 10 attempts
			if attempt > 10 {
				backoffDuration := time.Duration(math.Min(float64(attempt-5), 30)) * time.Second
				ticker.Reset(backoffDuration)
			}
		}
	}

	// If we reach here, we've exceeded our maximum attempts
	return nil, fmt.Errorf("exceeded maximum polling attempts (%d) waiting for transaction", maxAttempts)

TransactionComplete:
	receipt, err := node.TransactionReceipt(ctx, *txHash)

	if err != nil {
		return nil, err
	}

	if receipt == nil {
		return nil, fmt.Errorf("nil transaction receipt returned")
	}

	if len(receipt.Logs) == 0 {
		return nil, fmt.Errorf("no logs found in transaction receipt")
	}

	// Safely access the first log
	firstLog := receipt.Logs[0]
	if firstLog == nil {
		return nil, fmt.Errorf("nil log entry in transaction receipt")
	}

	if len(firstLog.Topics) < 2 {
		return nil, fmt.Errorf("invalid log topics length, expected at least 2, got %d", len(firstLog.Topics))
	}

	// Safely access the second topic
	if firstLog.Topics[1] == (common.Hash{}) {
		return nil, fmt.Errorf("empty topic hash at index 1")
	}

	splitAddress := common.HexToAddress(firstLog.Topics[1].Hex()).Hex()

	return &splitAddress, nil
}

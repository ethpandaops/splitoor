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

func (p *UpdateSplitParams) encode(splitAddress string) []interface{} {
	// Create pairs for sorting
	pairs := make([][2]interface{}, len(p.Accounts))
	for i := range p.Accounts {
		pairs[i] = [2]interface{}{p.Accounts[i], p.PercentageAllocations[i]}
	}

	// Sort by account address
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i][0].(string) < pairs[j][0].(string)
	})

	// Separate back into sorted slices
	accounts := make([]common.Address, len(p.Accounts))
	allocations := make([]uint32, len(p.PercentageAllocations))

	for i := range pairs {
		accounts[i] = common.HexToAddress(pairs[i][0].(string))

		allocation, ok := pairs[i][1].(uint32)
		if !ok {
			return nil
		}

		allocations[i] = allocation
	}

	return []interface{}{
		common.HexToAddress(splitAddress),
		accounts,
		allocations,
		p.DistributorFee,
	}
}

func (c *Client) Update(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI, from, privateKey string, gasLimit uint64, params *UpdateSplitParams) (*string, error) {
	if err := params.order(); err != nil {
		return nil, err
	}

	if c.splitAddress == nil {
		return nil, fmt.Errorf("split address is not set")
	}

	pKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, err
	}

	calldata, err := contractABI.EncodeMethodCalldata("updateSplit", params.encode(*c.splitAddress))
	if err != nil {
		return nil, err
	}

	txHash, err := node.WriteContract(ctx, c.contractAddress, calldata, from, pKey, big.NewInt(0), gasLimit)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	c.log.WithField("tx", *txHash).Info("Waiting for transaction to be included in a block")

	for {
		var isPending bool

		_, isPending, err = node.TransactionByHash(ctx, *txHash)
		if err != nil {
			return nil, err
		}

		if !isPending {
			break
		}

		time.Sleep(time.Second)
	}

	receipt, err := node.TransactionReceipt(ctx, *txHash)
	if err != nil {
		return nil, err
	}

	if len(receipt.Logs) == 0 {
		return nil, fmt.Errorf("no logs found in transaction receipt")
	}

	if len(receipt.Logs[0].Topics) < 2 {
		return nil, fmt.Errorf("invalid log topics length")
	}

	splitAddress := common.HexToAddress(receipt.Logs[0].Topics[1].Hex()).Hex()

	return &splitAddress, nil
}

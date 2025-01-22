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

func (p *DistributeETHParams) encode(splitAddress string) []interface{} {
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

	distributorAddress := p.DistributorAddress
	if distributorAddress == "" {
		distributorAddress = "0x0000000000000000000000000000000000000000"
	}

	return []interface{}{
		common.HexToAddress(splitAddress),
		accounts,
		allocations,
		p.DistributorFee,
		common.HexToAddress(distributorAddress),
	}
}

func (c *Client) DistributeETH(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI, from, privateKey string, gasLimit uint64, params *DistributeETHParams) error {
	if err := params.order(); err != nil {
		return err
	}

	if c.splitAddress == nil {
		return fmt.Errorf("split address is not set")
	}

	pKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return err
	}

	calldata, err := contractABI.EncodeMethodCalldata("distributeETH", params.encode(*c.splitAddress))
	if err != nil {
		return err
	}

	txHash, err := node.WriteContract(ctx, c.contractAddress, calldata, from, pKey, big.NewInt(0), gasLimit)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	c.log.WithField("tx", *txHash).Info("Waiting for transaction to be included in a block")

	for {
		var isPending bool

		_, isPending, err = node.TransactionByHash(ctx, *txHash)
		if err != nil {
			return err
		}

		if !isPending {
			break
		}

		time.Sleep(time.Second)
	}

	receipt, err := node.TransactionReceipt(ctx, *txHash)
	if err != nil {
		return err
	}

	if receipt.Status == types.ReceiptStatusFailed {
		return fmt.Errorf("transaction failed")
	}

	return nil
}

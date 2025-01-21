package split

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/0xsequence/ethkit/go-ethereum/crypto"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/holiman/uint256"
)

type WithdrawParams struct {
	Address     string
	WithdrawETH bool
	Tokens      []string
}

func (w *WithdrawParams) encode() []interface{} {
	withdrawETH := uint256.NewInt(0)
	if w.WithdrawETH {
		withdrawETH = uint256.NewInt(1)
	}

	tokens := make([]common.Address, len(w.Tokens))
	for i := range w.Tokens {
		tokens[i] = common.HexToAddress(w.Tokens[i])
	}

	return []interface{}{
		common.HexToAddress(w.Address),
		withdrawETH.ToBig(),
		tokens,
	}
}

func (c *Client) Withdraw(ctx context.Context, node *execution.Node, contractABI *ethcoder.ABI, from, privateKey string, gasLimit uint64, params *WithdrawParams) error {
	pKey, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return err
	}

	calldata, err := contractABI.EncodeMethodCalldata("withdraw", params.encode())
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

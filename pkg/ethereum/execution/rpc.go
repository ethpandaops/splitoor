package execution

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math"
	"math/big"

	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/0xsequence/ethkit/go-ethereum/crypto"
)

func (n *Node) BlockNumber(ctx context.Context) (*uint64, error) {
	var blockNumber uint64

	_, err := n.rpc.Do(ctx, ethrpc.BlockNumber().Into(&blockNumber))
	if err != nil {
		return nil, err
	}

	return &blockNumber, nil
}

func convertBlockNumber(blockNumber uint64) int64 {
	if blockNumber > uint64(math.MaxInt64) {
		return int64(math.MaxInt64)
	}

	return int64(blockNumber)
}

func (n *Node) NonceAt(ctx context.Context, address string) (*uint64, error) {
	blockNumber, err := n.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	var nonce uint64

	_, err = n.rpc.Do(ctx, ethrpc.NonceAt(common.HexToAddress(address), big.NewInt(convertBlockNumber(*blockNumber))).Into(&nonce))
	if err != nil {
		return nil, err
	}

	return &nonce, nil
}

func (n *Node) ChainID(ctx context.Context) (*big.Int, error) {
	var chainID *big.Int

	_, err := n.rpc.Do(ctx, ethrpc.ChainID().Into(&chainID))
	if err != nil {
		return nil, err
	}

	return chainID, nil
}

func (n *Node) BalanceAt(ctx context.Context, address string) (*big.Int, error) {
	blockNumber, err := n.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}

	var balance *big.Int

	_, err = n.rpc.Do(ctx, ethrpc.BalanceAt(common.HexToAddress(address), big.NewInt(convertBlockNumber(*blockNumber))).Into(&balance))
	if err != nil {
		return nil, err
	}

	return balance, nil
}

func (n *Node) SignTransaction(ctx context.Context, tx *types.Transaction, key *ecdsa.PrivateKey) (*types.Transaction, error) {
	chainID, err := n.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	signer := types.NewCancunSigner(chainID)

	signature, err := crypto.Sign(signer.Hash(tx).Bytes(), key)
	if err != nil {
		return nil, err
	}

	signTx, err := tx.WithSignature(signer, signature)
	if err != nil {
		return nil, err
	}

	return signTx, nil
}

func (n *Node) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	_, err := n.rpc.Do(ctx, ethrpc.SendTransaction(tx))
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) FeeHistory(ctx context.Context, blockCount uint64, lastBlock *big.Int, rewardPercentiles []float64) (*ethereum.FeeHistory, error) {
	var feeHistory *ethereum.FeeHistory

	_, err := n.rpc.Do(ctx, ethrpc.FeeHistory(blockCount, lastBlock, rewardPercentiles).Into(&feeHistory))
	if err != nil {
		return nil, err
	}

	return feeHistory, nil
}

func (n *Node) SuggestFees(ctx context.Context) (gasTipCap, gasFeeCap *big.Int, err error) {
	feeHistory, err := n.FeeHistory(ctx, 5, nil, []float64{0.50})
	if err != nil {
		return nil, nil, err
	}

	if len(feeHistory.BaseFee) < 2 || len(feeHistory.Reward) == 0 {
		return nil, nil, errors.New("invalid FeeHistory data")
	}

	baseFees := make([]*big.Int, len(feeHistory.Reward))
	tips := make([]*big.Int, len(feeHistory.Reward))

	for i := range feeHistory.Reward {
		baseFees[i] = feeHistory.BaseFee[i]
		tips[i] = feeHistory.Reward[i][0]
	}

	avgBaseFee := averageBigInts(baseFees)
	avgTip := averageBigInts(tips)

	// Project base fee with 12.5% increase
	projectedBaseFee := new(big.Int).Mul(avgBaseFee, big.NewInt(1125))
	projectedBaseFee.Div(projectedBaseFee, big.NewInt(1000))

	return avgTip, new(big.Int).Add(projectedBaseFee, avgTip), nil
}

func (n *Node) DeployContract(ctx context.Context, contract *[]byte, from string, key *ecdsa.PrivateKey, gasLimit uint64) (*string, error) {
	nonce, err := n.NonceAt(ctx, from)
	if err != nil {
		return nil, err
	}

	chainID, err := n.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	gasTipCap, gasFeeCap, err := n.SuggestFees(ctx)
	if err != nil {
		return nil, err
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     *nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        nil,
		Value:     big.NewInt(0),
		Data:      *contract,
	})

	signTx, err := n.SignTransaction(ctx, tx, key)
	if err != nil {
		return nil, err
	}

	txHash := signTx.Hash().Hex()

	err = n.SendTransaction(ctx, signTx)
	if err != nil {
		return nil, err
	}

	return &txHash, nil
}

func (n *Node) TransactionByHash(ctx context.Context, hash string) (*types.Transaction, bool, error) {
	var tx *types.Transaction

	var isPending bool

	_, err := n.rpc.Do(ctx, ethrpc.TransactionByHash(common.HexToHash(hash)).Into(&tx, &isPending))
	if err != nil {
		return nil, false, err
	}

	return tx, isPending, nil
}

func (n *Node) TransactionReceipt(ctx context.Context, hash string) (*types.Receipt, error) {
	var receipt *types.Receipt

	_, err := n.rpc.Do(ctx, ethrpc.TransactionReceipt(common.HexToHash(hash)).Into(&receipt))
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (n *Node) ReadContract(ctx context.Context, contractAddress string, callData []byte, blockNumber *big.Int) ([]byte, error) {
	var result []byte

	address := common.HexToAddress(contractAddress)

	msg := ethereum.CallMsg{
		To:   &address,
		Data: callData,
	}

	_, err := n.rpc.Do(ctx, ethrpc.CallContract(msg, blockNumber).Into(&result))
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (n *Node) WriteContract(ctx context.Context, contractAddress string, callData []byte, from string, key *ecdsa.PrivateKey, value *big.Int, gasLimit uint64) (*string, error) {
	nonce, err := n.NonceAt(ctx, from)
	if err != nil {
		return nil, err
	}

	chainID, err := n.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	gasTipCap, gasFeeCap, err := n.SuggestFees(ctx)
	if err != nil {
		return nil, err
	}

	to := common.HexToAddress(contractAddress)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     *nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     value,
		Data:      callData,
	})

	signTx, err := n.SignTransaction(ctx, tx, key)
	if err != nil {
		return nil, err
	}

	txHash := signTx.Hash().Hex()

	err = n.SendTransaction(ctx, signTx)
	if err != nil {
		return nil, err
	}

	return &txHash, nil
}

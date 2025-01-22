package cmd

import (
	"context"
	"strings"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	"github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	withdrawElRPCURL        string
	withdrawPrivKey         string
	withdrawContractAddress string
	withdrawAddress         string
	withdrawETH             bool
	withdrawTokens          string
	withdrawGasLimit        uint64
)

var withdrawSplitCmd = &cobra.Command{
	Use:   "withdraw",
	Short: "Withdraw from a split",
	Long:  `Withdraw ETH and/or tokens from a split.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := withdrawSplit(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	splitCmd.AddCommand(withdrawSplitCmd)

	withdrawSplitCmd.Flags().StringVar(&withdrawElRPCURL, "el-rpc-url", "", "Execution layer (EL) RPC URL")
	withdrawSplitCmd.Flags().StringVar(&withdrawPrivKey, "private-key", "", "Private key")
	withdrawSplitCmd.Flags().StringVar(&withdrawContractAddress, "contract", "", "Contract address")
	withdrawSplitCmd.Flags().StringVar(&withdrawAddress, "address", "", "Address to withdraw from")
	withdrawSplitCmd.Flags().BoolVar(&withdrawETH, "withdraw-eth", false, "Withdraw ETH, false if only withdrawing ERC20s")
	withdrawSplitCmd.Flags().StringVar(&withdrawTokens, "tokens", "", "Comma-separated list of ERC20 token addresses to withdraw")
	withdrawSplitCmd.Flags().Uint64Var(&withdrawGasLimit, "gaslimit", 3000000, "Gas limit for transaction")

	err := withdrawSplitCmd.MarkFlagRequired("el-rpc-url")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "el-rpc-url")
	}

	err = withdrawSplitCmd.MarkFlagRequired("private-key")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "private-key")
	}

	err = withdrawSplitCmd.MarkFlagRequired("address")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "address")
	}
}

func withdrawSplit(ctx context.Context) error {
	dpNodeConfig := &execution.Config{
		NodeAddress: withdrawElRPCURL,
	}

	dpNode := execution.NewNode(log, "execution", dpNodeConfig)
	if err := dpNode.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start execution node")
	}

	if withdrawContractAddress == "" {
		address, err := getSplitDefaultContractAddress(ctx, dpNode)
		if err != nil {
			return err
		}

		withdrawContractAddress = *address
	}

	config := &split.Config{
		ContractAddress: withdrawContractAddress,
	}

	client, err := split.NewClient(log, config)
	if err != nil {
		return err
	}

	contractABI, err := contract.GetSplitMainAbi()
	if err != nil {
		return err
	}

	var tokens []string
	if withdrawTokens != "" {
		tokens = strings.Split(withdrawTokens, ",")
	}

	params := &split.WithdrawParams{
		Address:     withdrawAddress,
		WithdrawETH: withdrawETH,
		Tokens:      tokens,
	}

	err = client.Withdraw(ctx, dpNode, contractABI, withdrawAddress, withdrawPrivKey, withdrawGasLimit, params)
	if err != nil {
		return err
	}

	log.Info("Split withdrawn")

	return nil
}

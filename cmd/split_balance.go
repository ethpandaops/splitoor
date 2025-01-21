package cmd

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	"github.com/ethpandaops/splitoor/pkg/0xsplits/split"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	balanceElRPCURL        string
	balanceContractAddress string
	balanceAddress         string
)

var balanceSplitCmd = &cobra.Command{
	Use:   "balance",
	Short: "Check distributed balance of an address",
	Long:  `Check the distributed balance of an address on the splits contract.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := getSplitBalance(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	splitCmd.AddCommand(balanceSplitCmd)

	balanceSplitCmd.Flags().StringVar(&balanceElRPCURL, "el-rpc-url", "", "Execution layer (EL) RPC URL")
	balanceSplitCmd.Flags().StringVar(&balanceContractAddress, "contract", "", "Contract address")
	balanceSplitCmd.Flags().StringVar(&balanceAddress, "address", "", "Address to check balance")

	err := balanceSplitCmd.MarkFlagRequired("el-rpc-url")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "el-rpc-url")
	}

	err = balanceSplitCmd.MarkFlagRequired("address")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "address")
	}
}

func getSplitBalance(ctx context.Context) error {
	dpNodeConfig := &execution.Config{
		NodeAddress: balanceElRPCURL,
	}

	dpNode := execution.NewNode(log, "execution", dpNodeConfig)
	if err := dpNode.Start(ctx); err != nil {
		return errors.Wrap(err, "failed to start execution node")
	}

	if balanceContractAddress == "" {
		address, err := getSplitDefaultContractAddress(ctx, dpNode)
		if err != nil {
			return err
		}

		balanceContractAddress = *address
	}

	config := &split.Config{
		ContractAddress: balanceContractAddress,
		SplitAddress:    &balanceAddress,
	}

	client, err := split.NewClient(log, config)
	if err != nil {
		return err
	}

	contractABI, err := contract.GetSplitMainAbi()
	if err != nil {
		return err
	}

	balance, err := client.GetETHBalance(ctx, dpNode, contractABI, balanceAddress)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"balance": balance,
		"address": balanceAddress,
	}).Info("Distributed balance")

	return nil
}

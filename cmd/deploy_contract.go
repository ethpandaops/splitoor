package cmd

import (
	"context"

	"github.com/ethpandaops/splitoor/pkg/0xsplits/contract"
	"github.com/ethpandaops/splitoor/pkg/ethereum/execution"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	dpElRPCURL        string
	dpDeployerAddress string
	dpPrivateKey      string
	dpGasLimit        uint64
)

var deployContractCmd = &cobra.Command{
	Use:   "deploy-contract",
	Short: "Deploy the contract",
	Long:  `Deploy the contract to the specified network.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		err := deployContract(cmd.Context())
		if err != nil {
			log.Fatal(err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deployContractCmd)

	deployContractCmd.Flags().StringVar(&dpElRPCURL, "el-rpc-url", "", "Execution layer (EL) RPC URL")
	deployContractCmd.Flags().StringVar(&dpDeployerAddress, "deployer-address", "", "Deployer address")
	deployContractCmd.Flags().StringVar(&dpPrivateKey, "deployer-private-key", "", "Deployer private key")
	deployContractCmd.Flags().Uint64Var(&dpGasLimit, "gaslimit", 3000000, "Gas limit for transaction")

	err := deployContractCmd.MarkFlagRequired("el-rpc-url")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "el-rpc-url")
	}

	err = deployContractCmd.MarkFlagRequired("deployer-address")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-address")
	}

	err = deployContractCmd.MarkFlagRequired("deployer-private-key")
	if err != nil {
		log.WithError(err).Fatalf("Failed to mark flag %s as required", "deployer-private-key")
	}
}

func deployContract(ctx context.Context) error {
	dpNodeConfig := &execution.Config{
		NodeAddress: dpElRPCURL,
	}

	dpNode := execution.NewNode(log, "execution", dpNodeConfig)
	if err := dpNode.Start(ctx); err != nil {
		return err
	}

	config := &contract.Config{
		From:       dpDeployerAddress,
		PrivateKey: dpPrivateKey,
		GasLimit:   dpGasLimit,
	}

	deployer, err := contract.NewDeployer(ctx, log, config)
	if err != nil {
		return err
	}

	contractAddress, err := deployer.Deploy(ctx, dpNode)
	if err != nil {
		return err
	}

	log.WithFields(logrus.Fields{
		"address": *contractAddress,
	}).Info("Contract deployed")

	return nil
}

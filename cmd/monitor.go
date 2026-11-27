package cmd

import (
	"context"

	"github.com/creasty/defaults"
	m "github.com/ethpandaops/splitoor/pkg/monitor"
	"github.com/spf13/cobra"
)

var monitorConfigFile string

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor",
	Long:  `Monitor splits`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initCommon()

		monitor(cmd.Context())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().StringVar(&monitorConfigFile, "config", "config.yaml", "Config file (default is config.yaml)")
}

func monitor(ctx context.Context) {
	config, err := loadMonitorConfigFromFile(monitorConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	server, err := m.NewServer(ctx, log, config)
	if err != nil {
		log.Fatal(err)
	}

	if startErr := server.Start(ctx); startErr != nil {
		log.Fatal(startErr)
	}

	log.Info("Splitoor monitor exited - cya!")
}

func loadMonitorConfigFromFile(file string) (*m.Config, error) {
	if file == "" {
		file = "config.yaml"
	}

	config := &m.Config{}

	if defaultsErr := defaults.Set(config); defaultsErr != nil {
		return nil, defaultsErr
	}

	type plain m.Config

	if err := loadConfigFile(file, (*plain)(config)); err != nil {
		return nil, err
	}

	return config, nil
}

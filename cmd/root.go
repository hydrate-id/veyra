package cmd

import (
	"fmt"
	"os"

	"github.com/korfra/veyra/core"
	"github.com/korfra/veyra/store"
	"github.com/spf13/cobra"
)

var dbPath string
var verbose bool
var osExit = os.Exit

var RootCmd = &cobra.Command{
	Use:   "veyra",
	Short: "Veyra is a local reverse proxy control plane using embedded Caddy",
	Long: `Veyra is a single-binary local ingress controller based on Caddy that manages
domain mapping, reverse proxy rules, and dynamic config reloading.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		osExit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&dbPath, "db", "veyra.db", "path to BoltDB file")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func InitRegistry() (*core.Registry, func(), error) {
	s, err := store.New(dbPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w (is another instance running?)", err)
	}
	reg := core.New(s)
	cleanup := func() {
		_ = s.Close()
	}
	return reg, cleanup, nil
}

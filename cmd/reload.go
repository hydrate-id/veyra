package cmd

import (
	"os"

	"github.com/korfra/veyra/caddy"
	"github.com/spf13/cobra"
)

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reload the Caddy configuration manually",
	Run: func(cmd *cobra.Command, args []string) {
		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			os.Exit(1)
		}

		caddyfileContent, err := reg.GenerateCaddyfile()
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error generating Caddyfile: %v\n", err)
			os.Exit(1)
		}

		adminPort, _ := reg.GetAdminPort()
		cleanup()

		err = caddy.Reload(caddyfileContent, adminPort)
		if err != nil {
			cmd.PrintErrf("Error reloading Caddy: %v\n", err)
			os.Exit(1)
		}

		cmd.Println("Caddy configuration reloaded successfully.")
	},
}

func init() {
	RootCmd.AddCommand(reloadCmd)
}

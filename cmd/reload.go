package cmd

import (
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
			osExit(1)
		}

		caddyfileContent, err := reg.GenerateCaddyfile()
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error generating Caddyfile: %v\n", err)
			osExit(1)
		}

		adminPort, _ := reg.GetAdminPort()
		cleanup()

		err = caddy.Reload(caddyfileContent, adminPort)
		if err != nil {
			if verbose {
				cmd.PrintErrf("Error reloading Caddy: %v\n", err)
			} else {
				cmd.PrintErrf("Error: Could not reload Caddy configuration (is 'veyra start' running?)\n")
			}
			osExit(1)
		}

		cmd.Println("Caddy configuration reloaded successfully.")
	},
}

func init() {
	RootCmd.AddCommand(reloadCmd)
}

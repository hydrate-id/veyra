package cmd

import (
	"os"

	"github.com/korfra/veyra/caddy"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <public>",
	Short: "Remove a domain routing rule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		public := args[0]

		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			os.Exit(1)
		}

		existing, err := reg.Get(public)
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error checking rule: %v\n", err)
			os.Exit(1)
		}
		if existing == nil {
			cleanup()
			cmd.PrintErrf("Error: rule for '%s' not found\n", public)
			os.Exit(1)
		}

		err = reg.Remove(public)
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error removing rule: %v\n", err)
			os.Exit(1)
		}

		cmd.Printf("Removed rule for public pattern: %s\n", public)

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
			cmd.Printf("Warning: Could not reload Caddy configuration: %v\n", err)
			cmd.Println("Start the veyra daemon ('veyra start') to apply the rules.")
		} else {
			cmd.Println("Caddy configuration reloaded successfully.")
		}
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}

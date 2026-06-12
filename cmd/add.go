package cmd

import (
	"os"

	"github.com/korfra/veyra/caddy"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <public> <internal>",
	Short: "Add a new domain routing rule",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		public := args[0]
		internal := args[1]

		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			os.Exit(1)
		}

		err = reg.Add(public, internal)
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error adding rule: %v\n", err)
			os.Exit(1)
		}

		cmd.Printf("Added rule: %s -> %s\n", public, internal)

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
	RootCmd.AddCommand(addCmd)
}

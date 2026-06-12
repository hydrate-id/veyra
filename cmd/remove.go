package cmd

import (
	"fmt"
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
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		existing, err := reg.Get(public)
		if err != nil {
			cleanup()
			fmt.Fprintf(os.Stderr, "Error checking rule: %v\n", err)
			os.Exit(1)
		}
		if existing == nil {
			cleanup()
			fmt.Fprintf(os.Stderr, "Error: rule for '%s' not found\n", public)
			os.Exit(1)
		}

		err = reg.Remove(public)
		if err != nil {
			cleanup()
			fmt.Fprintf(os.Stderr, "Error removing rule: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Removed rule for public pattern: %s\n", public)

		caddyfileContent, err := reg.GenerateCaddyfile()
		if err != nil {
			cleanup()
			fmt.Fprintf(os.Stderr, "Error generating Caddyfile: %v\n", err)
			os.Exit(1)
		}

		cleanup()

		err = caddy.Reload(caddyfileContent)
		if err != nil {
			fmt.Printf("Warning: Could not reload Caddy configuration: %v\n", err)
			fmt.Println("Start the veyra daemon ('veyra start') to apply the rules.")
		} else {
			fmt.Println("Caddy configuration reloaded successfully.")
		}
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}

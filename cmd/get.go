package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <public>",
	Short: "Get details of a single rule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		public := args[0]

		reg, cleanup, err := InitRegistry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		rule, err := reg.Get(public)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting rule: %v\n", err)
			os.Exit(1)
		}

		if rule == nil {
			fmt.Fprintf(os.Stderr, "Error: rule for '%s' not found\n", public)
			os.Exit(1)
		}

		fmt.Printf("Type: static\n")
		fmt.Printf("Public:\n  %s\n\n", rule.PublicPattern)
		fmt.Printf("Internal:\n  %s\n", rule.InternalTarget)
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}

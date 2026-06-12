package cmd

import (
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
			cmd.PrintErrf("Error: %v\n", err)
			osExit(1)
		}
		defer cleanup()

		rule, err := reg.Get(public)
		if err != nil {
			cmd.PrintErrf("Error getting rule: %v\n", err)
			osExit(1)
		}

		if rule == nil {
			cmd.PrintErrf("Error: rule for '%s' not found\n", public)
			osExit(1)
		}

		cmd.Printf("Type: static\n")
		cmd.Printf("Public:\n  %s\n\n", rule.PublicPattern)
		cmd.Printf("Internal:\n  %s\n", rule.InternalTarget)
	},
}

func init() {
	RootCmd.AddCommand(getCmd)
}

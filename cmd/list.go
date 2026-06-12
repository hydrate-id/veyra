package cmd

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domain routing rules",
	Run: func(cmd *cobra.Command, args []string) {
		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			osExit(1)
		}
		defer cleanup()

		rules, err := reg.List()
		if err != nil {
			cmd.PrintErrf("Error listing rules: %v\n", err)
			osExit(1)
		}

		if len(rules) == 0 {
			cmd.Println("No rules registered.")
			return
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 4, ' ', 0)
		for _, rule := range rules {
			fmt.Fprintf(w, "%s\t->\t%s\n", rule.PublicPattern, rule.InternalTarget)
		}
		w.Flush()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}

package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all domain routing rules",
	Run: func(cmd *cobra.Command, args []string) {
		reg, cleanup, err := InitRegistry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		rules, err := reg.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing rules: %v\n", err)
			os.Exit(1)
		}

		if len(rules) == 0 {
			fmt.Println("No rules registered.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		for _, rule := range rules {
			fmt.Fprintf(w, "%s\t->\t%s\n", rule.PublicPattern, rule.InternalTarget)
		}
		w.Flush()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}

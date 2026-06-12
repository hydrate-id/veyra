package cmd

import (
	"github.com/korfra/veyra/caddy"
	"github.com/korfra/veyra/core"
	"github.com/korfra/veyra/model"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate all registered routing rules and generated Caddyfile config",
	Run: func(cmd *cobra.Command, args []string) {
		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			osExit(1)
		}

		rules, err := reg.List()
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error listing rules: %v\n", err)
			osExit(1)
		}

		for i, rule := range rules {
			otherRules := make([]model.Rule, 0, len(rules)-1)
			for j, other := range rules {
				if i != j {
					otherRules = append(otherRules, other)
				}
			}

			if err := core.ValidateRule(rule, otherRules); err != nil {
				cleanup()
				cmd.PrintErrf("Rule validation failed for pattern '%s': %v\n", rule.PublicPattern, err)
				osExit(1)
			}
		}

		caddyfileContent, err := reg.GenerateCaddyfile()
		if err != nil {
			cleanup()
			cmd.PrintErrf("Caddyfile generation failed: %v\n", err)
			osExit(1)
		}

		cleanup()

		_, err = caddy.AdaptCaddyfile(caddyfileContent)
		if err != nil {
			cmd.PrintErrf("Caddyfile adaptation failed: %v\n", err)
			osExit(1)
		}

		cmd.Println("All rules and Caddyfile configuration are valid.")
	},
}

func init() {
	RootCmd.AddCommand(validateCmd)
}

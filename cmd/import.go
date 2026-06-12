package cmd

import (
	"encoding/json"
	"os"

	"github.com/korfra/veyra/caddy"
	"github.com/korfra/veyra/core"
	"github.com/korfra/veyra/model"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <format> <file_path>",
	Short: "Import configuration (formats: json, backup)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		format := args[0]
		filePath := args[1]

		data, err := os.ReadFile(filePath)
		if err != nil {
			cmd.PrintErrf("Error reading file '%s': %v\n", filePath, err)
			os.Exit(1)
		}

		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			os.Exit(1)
		}

		switch format {
		case "json":
			var rules []model.Rule
			err = json.Unmarshal(data, &rules)
			if err != nil {
				cleanup()
				cmd.PrintErrf("Error parsing rules JSON: %v\n", err)
				os.Exit(1)
			}
			for _, rule := range rules {
				err = reg.Add(rule.PublicPattern, rule.InternalTarget)
				if err != nil {
					cmd.Printf("Warning: Failed to import rule '%s -> %s': %v\n", rule.PublicPattern, rule.InternalTarget, err)
				} else {
					cmd.Printf("Imported rule: %s -> %s\n", rule.PublicPattern, rule.InternalTarget)
				}
			}

		case "backup":
			var backup core.BackupData
			err = json.Unmarshal(data, &backup)
			if err != nil {
				cleanup()
				cmd.PrintErrf("Error parsing backup JSON: %v\n", err)
				os.Exit(1)
			}
			err = reg.ImportBackup(&backup)
			if err != nil {
				cleanup()
				cmd.PrintErrf("Error importing backup: %v\n", err)
				os.Exit(1)
			}
			cmd.Println("Backup configuration imported successfully.")

		default:
			cleanup()
			cmd.PrintErrf("Error: Invalid format '%s'. Valid formats are: json, backup\n", format)
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
			cmd.Printf("Warning: Could not reload Caddy configuration: %v\n", err)
			cmd.Println("Start the veyra daemon ('veyra start') to apply the rules.")
		} else {
			cmd.Println("Caddy configuration reloaded successfully.")
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)
}

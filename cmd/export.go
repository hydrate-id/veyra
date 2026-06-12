package cmd

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export <format> [file_path]",
	Short: "Export configuration (formats: caddy, json, backup)",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		format := args[0]
		var filePath string
		if len(args) > 1 {
			filePath = args[1]
		}

		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			osExit(1)
		}
		defer cleanup()

		var output []byte

		switch format {
		case "caddy":
			caddyfileContent, err := reg.GenerateCaddyfile()
			if err != nil {
				cmd.PrintErrf("Error generating Caddyfile: %v\n", err)
				osExit(1)
			}
			output = []byte(caddyfileContent)

		case "json":
			rules, err := reg.List()
			if err != nil {
				cmd.PrintErrf("Error listing rules: %v\n", err)
				osExit(1)
			}
			data, err := json.MarshalIndent(rules, "", "  ")
			if err != nil {
				cmd.PrintErrf("Error marshaling rules to JSON: %v\n", err)
				osExit(1)
			}
			output = data

		case "backup":
			backup, err := reg.ExportBackup()
			if err != nil {
				cmd.PrintErrf("Error exporting backup: %v\n", err)
				osExit(1)
			}
			data, err := json.MarshalIndent(backup, "", "  ")
			if err != nil {
				cmd.PrintErrf("Error marshaling backup to JSON: %v\n", err)
				osExit(1)
			}
			output = data

		default:
			cmd.PrintErrf("Error: Invalid format '%s'. Valid formats are: caddy, json, backup\n", format)
			osExit(1)
		}

		if filePath != "" {
			err = os.WriteFile(filePath, output, 0644)
			if err != nil {
				cmd.PrintErrf("Error writing to file '%s': %v\n", filePath, err)
				osExit(1)
			}
			cmd.Printf("Configuration exported successfully to '%s'\n", filePath)
		} else {
			cmd.Print(string(output))
		}
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)
}

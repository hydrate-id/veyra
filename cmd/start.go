package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/korfra/veyra/caddy"
	"github.com/spf13/cobra"
)

var httpPort int
var httpsPort int
var adminPort int
var httpsEnabled bool

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Veyra daemon and embedded Caddy instance",
	Run: func(cmd *cobra.Command, args []string) {
		reg, cleanup, err := InitRegistry()
		if err != nil {
			cmd.PrintErrf("Error: %v\n", err)
			osExit(1)
		}

		err = reg.SetPorts(httpPort, httpsPort)
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error saving port settings: %v\n", err)
			osExit(1)
		}

		err = reg.SetAdminPort(adminPort)
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error saving admin port settings: %v\n", err)
			osExit(1)
		}

		err = reg.SetHTTPSEnabled(httpsEnabled)
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error saving HTTPS setting: %v\n", err)
			osExit(1)
		}

		caddyfileContent, err := reg.GenerateCaddyfile()
		if err != nil {
			cleanup()
			cmd.PrintErrf("Error generating initial Caddyfile: %v\n", err)
			osExit(1)
		}

		cleanup()

		cmd.Println("Starting Veyra reverse proxy daemon...")
		if httpsEnabled {
			cmd.Printf("HTTP Port: %d, HTTPS Port: %d\n", httpPort, httpsPort)
		} else {
			cmd.Printf("HTTP Port: %d (HTTPS disabled)\n", httpPort)
		}
		cmd.Printf("Admin API: http://localhost:%d\n", adminPort)

		err = caddy.Start(caddyfileContent)
		if err != nil {
			cmd.PrintErrf("Error starting embedded Caddy: %v\n", err)
			osExit(1)
		}

		cmd.Println("Veyra is running. Press Ctrl+C to stop.")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		cmd.Println("\nStopping Veyra daemon and cleaning up...")
		if err := caddy.Stop(); err != nil {
			cmd.PrintErrf("Error stopping Caddy: %v\n", err)
		}
		cmd.Println("Veyra daemon stopped.")
	},
}

func init() {
	startCmd.Flags().IntVar(&httpPort, "http-port", 8080, "HTTP port for Caddy")
	startCmd.Flags().IntVar(&httpsPort, "https-port", 8443, "HTTPS port for Caddy")
	startCmd.Flags().IntVar(&adminPort, "admin-port", 2019, "Admin API port for Caddy")
	startCmd.Flags().BoolVar(&httpsEnabled, "https", false, "enable HTTPS")
	RootCmd.AddCommand(startCmd)
}

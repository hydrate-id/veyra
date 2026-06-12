package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/korfra/veyra/caddy"
	"github.com/spf13/cobra"
)

var httpPort int
var httpsPort int

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Veyra daemon and embedded Caddy instance",
	Run: func(cmd *cobra.Command, args []string) {
		reg, cleanup, err := InitRegistry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		err = reg.SetPorts(httpPort, httpsPort)
		if err != nil {
			cleanup()
			fmt.Fprintf(os.Stderr, "Error saving port settings: %v\n", err)
			os.Exit(1)
		}

		caddyfileContent, err := reg.GenerateCaddyfile()
		if err != nil {
			cleanup()
			fmt.Fprintf(os.Stderr, "Error generating initial Caddyfile: %v\n", err)
			os.Exit(1)
		}

		cleanup()

		fmt.Println("Starting Veyra reverse proxy daemon...")
		fmt.Printf("HTTP Port: %d, HTTPS Port: %d\n", httpPort, httpsPort)
		fmt.Println("Admin API: http://localhost:2019")

		err = caddy.Start(caddyfileContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting embedded Caddy: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Veyra is running. Press Ctrl+C to stop.")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\nStopping Veyra daemon and cleaning up...")
		if err := caddy.Stop(); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping Caddy: %v\n", err)
		}
		fmt.Println("Veyra daemon stopped.")
	},
}

func init() {
	startCmd.Flags().IntVar(&httpPort, "http-port", 80, "HTTP port for Caddy")
	startCmd.Flags().IntVar(&httpsPort, "https-port", 443, "HTTPS port for Caddy")
	RootCmd.AddCommand(startCmd)
}

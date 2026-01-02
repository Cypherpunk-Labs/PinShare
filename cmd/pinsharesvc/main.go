package main

import (
	"fmt"
	"log"
	"os"

	"pinshare/internal/winservice"

	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
)

var rootCmd = &cobra.Command{
	Use:   "pinsharesvc",
	Short: "PinShare Windows Service",
	Long:  "PinShare Windows Service wrapper - manages IPFS and PinShare backend as a Windows service.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Check if running as Windows service before any command
		isWindowsService, err := svc.IsWindowsService()
		if err != nil {
			log.Fatalf("Failed to determine if running as service: %v", err)
		}

		if isWindowsService {
			// Run as Windows service and exit
			runService()
			os.Exit(0)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand provided, show usage
		cmd.Help()
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install PinShare as a Windows service",
	Long:  "Install PinShare as a Windows service. Use --auto-start to start automatically on boot.",
	RunE: func(cmd *cobra.Command, args []string) error {
		autoStart, _ := cmd.Flags().GetBool("auto-start")
		if err := installService(autoStart); err != nil {
			return err
		}
		fmt.Println("Successfully installed PinShare service")
		return nil
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall PinShare Windows service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := uninstallService(); err != nil {
			return err
		}
		fmt.Println("Successfully uninstalled PinShare service")
		return nil
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start PinShare service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := startService(); err != nil {
			return err
		}
		fmt.Println("Successfully started PinShare service")
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop PinShare service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := stopService(); err != nil {
			return err
		}
		fmt.Println("Successfully stopped PinShare service")
		return nil
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart PinShare service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := restartService(); err != nil {
			return err
		}
		fmt.Println("Successfully restarted PinShare service")
		return nil
	},
}

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Run in console mode for debugging",
	Long:  "Run PinShare in console mode for debugging. Press Ctrl+C to stop.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running PinShare in debug mode...")
		fmt.Println("Press Ctrl+C to stop")
		return runDebugMode()
	},
}

func init() {
	// Add --auto-start flag to install command
	installCmd.Flags().Bool("auto-start", false, "Start service automatically on boot")

	// Register all subcommands
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(debugCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runService() {
	err := svc.Run(winservice.ServiceName, new(pinshareService))
	if err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}

func runDebugMode() error {
	service := new(pinshareService)
	return service.runInteractive()
}

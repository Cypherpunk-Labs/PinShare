package cmd

import (
	"pinshare/internal/p2p" // Import for SetP2PManager

	"github.com/spf13/cobra"
)

// Global P2PManager accessible by commands if needed, set by main
var p2pManagerInstance *p2p.PubSubManager

var rootCmd = &cobra.Command{
	Use:   "metadata-manager",
	Short: "A CLI tool to manage decentralized file metadata",
	Long: `Metadata Manager CLI is a proof-of-concept tool to interact with
a local metadata store, designed for a decentralized file sharing system.
It also initializes a libp2p host for peer-to-peer interactions.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// This function can be used to initialize things before any command runs.
		// For example, ensuring the libp2p host is ready if commands need it.
		// fmt.Println("[DEBUG] PersistentPreRun called")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// SetP2PManager allows main to set the global PubSubManager instance
func SetP2PManager(manager *p2p.PubSubManager) {
	p2pManagerInstance = manager
}

func init() {
	// cobra.OnInitialize(initConfig) // Example for config file loading
}

// Helper function to get the data file path, can be made configurable later
func getDataFilePath() string {
	// For POC, hardcode. Could use env var or flag.
	return "metadata.json"
}

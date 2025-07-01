package cmd

// import (
// 	"encoding/json"
// 	"fmt"

// 	"pinshare/internal/store"

// 	"github.com/spf13/cobra"
// )

// var getCmd = &cobra.Command{
// 	Use:   "get <fileSHA256>",
// 	Short: "Get metadata for a specific file",
// 	Long:  `Retrieves and displays the metadata for a file identified by its SHA256 hash.`,
// 	Args:  cobra.ExactArgs(1),
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		fileSHA256 := args[0]
// 		metadata, exists := store.GlobalStore.GetFile(fileSHA256)
// 		if !exists {
// 			return fmt.Errorf("no metadata found for fileSHA256: %s", fileSHA256)
// 		}

// 		jsonData, err := json.MarshalIndent(metadata, "", "  ")
// 		if err != nil {
// 			return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
// 		}

// 		fmt.Println(string(jsonData))
// 		return nil
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(getCmd)
// }

package cmd

// import (
// 	"fmt"
// 	"strings"

// 	"pinshare/internal/store"

// 	"github.com/spf13/cobra"
// )

// var (
// 	addIPFSCID  string
// 	addFileType string
// 	// addTitle           string
// 	// addAuthor          string
// 	// addFileName        string
// 	// addDate            string
// 	// addScientificField string
// 	// addTags            string // Comma-separated
// )

// var addCmd = &cobra.Command{
// 	Use:   "add <fileSHA256>",
// 	Short: "Add metadata for a new file",
// 	Long:  `Adds a new metadata entry to the store. FileSHA256 is required.`,
// 	Args:  cobra.ExactArgs(1), // Requires exactly one argument: fileSHA256
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		fileSHA256 := args[0]
// 		if fileSHA256 == "" {
// 			return fmt.Errorf("fileSHA256 cannot be empty")
// 		}
// 		if addIPFSCID == "" {
// 			fmt.Println("Warning: IPFS CID is empty. This is usually a key identifier.")
// 		}

// 		tagsMap := make(map[string]bool)
// 		if addTags != "" {
// 			tagsList := strings.Split(addTags, ",")
// 			for _, t := range tagsList {
// 				trimmedTag := strings.TrimSpace(t)
// 				if trimmedTag != "" {
// 					tagsMap[trimmedTag] = true
// 				}
// 			}
// 		}

// 		metadata := store.BaseMetadata{
// 			FileSHA256: fileSHA256,
// 			IPFSCID:    addIPFSCID,
// 			FileType:   addFileType,
// 			// Title:           addTitle,
// 			// Author:          addAuthor,
// 			// FileName:        addFileName,
// 			// Date:            addDate,
// 			// ScientificField: addScientificField,
// 			// Tags:            tagsMap,
// 			// CommunityLabels: make(map[string]int),
// 			// ModerationVotes: 0,
// 			// AddedAt and LastUpdated will be set by the store logic
// 		}

// 		err := store.GlobalStore.AddFile(metadata)
// 		if err != nil {
// 			return fmt.Errorf("failed to add file metadata: %w", err)
// 		}

// 		fmt.Printf("Successfully added/updated metadata for %s\n", fileSHA256)
// 		// Save the entire store to disk after this operation
// 		return store.GlobalStore.Save(getDataFilePath())
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(addCmd)

// 	addCmd.Flags().StringVarP(&addIPFSCID, "cid", "c", "", "IPFS CID of the file")
// 	addCmd.Flags().StringVarP(&addTitle, "title", "t", "", "Title of the file/paper")
// 	addCmd.Flags().StringVarP(&addAuthor, "author", "a", "", "Author(s) of the file/paper")
// 	addCmd.Flags().StringVarP(&addFileName, "filename", "f", "", "Original filename")
// 	addCmd.Flags().StringVarP(&addDate, "date", "d", "", "Publication date (e.g., YYYY-MM-DD)")
// 	addCmd.Flags().StringVarP(&addScientificField, "field", "s", "", "Scientific field")
// 	addCmd.Flags().StringVar(&addTags, "tags", "", "Comma-separated list of tags (e.g., 'tag1,tag2')")
// }

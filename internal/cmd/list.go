package cmd

// import (
// 	"encoding/json"
// 	"fmt"
// 	"pinshare/internal/store"
// 	"strings"

// 	"github.com/spf13/cobra"
// )

// var (
// 	listTagFilter    string
// 	listAuthorFilter string
// )

// var listCmd = &cobra.Command{
// 	Use:   "list",
// 	Short: "List all metadata entries",
// 	Long:  `Lists all stored metadata entries. Can be filtered by tag or author.`,
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		allMetadata := store.GlobalStore.GetAllFiles()

// 		if len(allMetadata) == 0 {
// 			fmt.Println("No metadata entries found.")
// 			return nil
// 		}

// 		var filteredMetadata []*store.BaseMetadata
// 		if listTagFilter == "" && listAuthorFilter == "" {
// 			filteredMetadata = allMetadata
// 		} else {
// 			for _, meta := range allMetadata {
// 				matchTag := true
// 				if listTagFilter != "" {
// 					matchTag = false
// 					if meta.Tags != nil {
// 						for tag := range meta.Tags {
// 							if strings.EqualFold(tag, listTagFilter) {
// 								matchTag = true
// 								break
// 							}
// 						}
// 					}
// 				}

// 				matchAuthor := true
// 				if listAuthorFilter != "" {
// 					matchAuthor = false
// 					if strings.Contains(strings.ToLower(meta.Author), strings.ToLower(listAuthorFilter)) {
// 						matchAuthor = true
// 					}
// 				}

// 				if matchTag && matchAuthor {
// 					filteredMetadata = append(filteredMetadata, meta)
// 				}
// 			}
// 		}

// 		if len(filteredMetadata) == 0 {
// 			fmt.Println("No metadata entries match the filters.")
// 			return nil
// 		}

// 		jsonData, err := json.MarshalIndent(filteredMetadata, "", "  ")
// 		if err != nil {
// 			return fmt.Errorf("failed to marshal metadata to JSON: %w", err)
// 		}

// 		fmt.Println(string(jsonData))
// 		return nil
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(listCmd)
// 	listCmd.Flags().StringVar(&listTagFilter, "tag", "", "Filter by a specific tag (case-insensitive)")
// 	listCmd.Flags().StringVar(&listAuthorFilter, "author", "", "Filter by author (case-insensitive, substring match)")
// }

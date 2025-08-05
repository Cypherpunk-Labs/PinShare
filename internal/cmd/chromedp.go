package cmd

import (
	"fmt"
	"pinshare/internal/psfs"

	"github.com/spf13/cobra"
)

var testcdpCmd = &cobra.Command{
	Use:   "testcdp <fileSHA256>",
	Short: "Test ChromeDP",
	Long:  `Test ChromeDP and get browser version.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD testcdp called")
		err := psfs.ChromedpTest()
		if err != nil {
			return err
		}
		return nil
	},
}

// moved chromedpTest() to sec.go

func init() {
	rootCmd.AddCommand(testcdpCmd)
}

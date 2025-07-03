package cmd

import (
	"fmt"
	"pinshare/internal/psfs"

	"github.com/spf13/cobra"
)

var testslCmd = &cobra.Command{
	Use:   "testsl <fileSHA256>",
	Short: "Test a specific file",
	Long:  `Tests a file identified by its SHA256 hash.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD testsl called")
		fileSHA256 := args[0]

		verdict, err := psfs.GetVirusTotalVerdictByHash(fileSHA256)

		if err != nil {
			return err
		}

		fmt.Printf("Verdict safe: %t\n", verdict)

		return nil
	},
}

var testssCmd = &cobra.Command{
	Use:   "testss <filepath>",
	Short: "Test submit file",
	Long:  `Test submission of a file`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD testss called")
		filepath := args[0]

		verdict, err := psfs.SendFileToVirusTotal(filepath)

		if err != nil {
			return err
		}

		fmt.Printf("Verdict safe: %t\n", verdict)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testslCmd)
	rootCmd.AddCommand(testssCmd)
}

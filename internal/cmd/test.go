package cmd

import (
	"fmt"
	"pinshare/internal/psfs"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test <fileSHA256>",
	Short: "Test a specific file",
	Long:  `Tests a file identified by its SHA256 hash.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD test called")
		fileSHA256 := args[0]

		verdict, err := psfs.GetVirusTotalVerdictByHash(fileSHA256)

		if err != nil {
			return err
		}

		fmt.Printf("Verdict: %t\n", verdict)

		return nil
	},
}

// var unittestCmd = &cobra.Command{
// 	Use:   "unittest",
// 	Short: "Test a specific file",
// 	Long:  `Tests a file identified by its SHA256 hash.`,
// 	Args:  cobra.ExactArgs(1),
// 	RunE: func(cmd *cobra.Command, args []string) error {
// 		// fileSHA256 := args[0]

// 		testing.
// 		psfs.Test_GetVirusTotalReportByHash(testing.T{})

// 		// verdict, err := psfs.GetVirusTotalVerdictByHash(fileSHA256)

// 		// if err != nil {
// 		// 	return err
// 		// }

// 		// fmt.Printf("Verdict: %t\n", verdict)

// 		return nil
// 	},
// }

func init() {
	rootCmd.AddCommand(testCmd)
}

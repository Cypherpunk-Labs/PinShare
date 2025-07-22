package cmd

import (
	"fmt"
	"os"
	"pinshare/internal/config"
	"pinshare/internal/p2p"
	"pinshare/internal/store"
	"strconv"

	"github.com/spf13/cobra"
)

var storeTagAddCmd = &cobra.Command{
	Use:   "store-tag-add <fileSHA256> <tag>",
	Short: "Tags a specific file",
	Long:  `Tags a file identified by its SHA256 hash.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD store-tag-add called")
		fileSHA256 := args[0]
		tag := args[1]

		appconf, _ := config.LoadConfig()
		p2p.SetAppConfig(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}

		store.GlobalStore.AddTag(fileSHA256, tag)

		if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
		}

		return nil
	},
}

var storeTagDelCmd = &cobra.Command{
	Use:   "store-tag-del <fileSHA256> <tag>",
	Short: "Untags a specific file",
	Long:  `Untags a file identified by its SHA256 hash.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD store-tag-del called")
		fileSHA256 := args[0]
		tag := args[1]

		appconf, _ := config.LoadConfig()
		p2p.SetAppConfig(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}

		store.GlobalStore.RemoveTag(fileSHA256, tag)

		if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
		}

		return nil
	},
}

var storeTagVoteDownCmd = &cobra.Command{
	Use:   "store-tag-vote-down <fileSHA256> <tag>",
	Short: "Votes for tag of a specific file",
	Long:  `Decrements tag for a file identified by its SHA256 hash.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD store-tag-vote-down called")
		fileSHA256 := args[0]
		tag := args[1]

		appconf, _ := config.LoadConfig()
		p2p.SetAppConfig(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}

		store.GlobalStore.VoteOnTag(fileSHA256, tag, false)

		if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
		}

		return nil
	},
}

var storeTagVoteUpCmd = &cobra.Command{
	Use:   "store-tag-vote-up <fileSHA256> <tag>",
	Short: "Votes for tag of a specific file",
	Long:  `Incements tag Vote for a file identified by its SHA256 hash.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD store-tag-vote-up called")
		fileSHA256 := args[0]
		tag := args[1]

		appconf, _ := config.LoadConfig()
		p2p.SetAppConfig(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}

		store.GlobalStore.VoteOnTag(fileSHA256, tag, true)

		if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
		}

		return nil
	},
}

var storeVoteUpCmd = &cobra.Command{
	Use:   "store-vote-up <fileSHA256>",
	Short: "up vote a file",
	Long:  `Increment the moderation vote count of a file`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD store-vote-up called")
		fileSHA256 := args[0]

		appconf, _ := config.LoadConfig()
		p2p.SetAppConfig(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}

		store.GlobalStore.VoteForRemoval(fileSHA256, true)

		if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
		}
		return nil
	},
}

var storeVoteDownCmd = &cobra.Command{
	Use:   "store-vote-down <fileSHA256>",
	Short: "down vote a file",
	Long:  `Decrement the moderation vote count of a file`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD store-vote-down called")
		fileSHA256 := args[0]

		appconf, _ := config.LoadConfig()
		p2p.SetAppConfig(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}

		store.GlobalStore.VoteForRemoval(fileSHA256, false)

		if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
		}
		return nil
	},
}

var storeBanCmd = &cobra.Command{
	Use:   "store-ban <fileSHA256> <value>",
	Short: "down vote a file",
	Long:  `Decrement the moderation vote count of a file`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD store-ban called")
		fileSHA256 := args[0]
		value := args[1] // convert to int
		intValue, err1 := strconv.Atoi(value)
		if err1 != nil {
			return fmt.Errorf("invalid value for ban reason: %w", err1)
		}

		appconf, _ := config.LoadConfig()
		p2p.SetAppConfig(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}

		store.GlobalStore.BanFile(fileSHA256, intValue)

		if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(storeTagAddCmd)
	rootCmd.AddCommand(storeTagDelCmd)
	rootCmd.AddCommand(storeVoteUpCmd)
	rootCmd.AddCommand(storeVoteDownCmd)
	rootCmd.AddCommand(storeTagVoteUpCmd)
	rootCmd.AddCommand(storeTagVoteDownCmd)
	rootCmd.AddCommand(storeBanCmd)
}

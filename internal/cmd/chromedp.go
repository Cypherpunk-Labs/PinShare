package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/chromedp/chromedp"

	"github.com/spf13/cobra"
)

var cdpCmd = &cobra.Command{
	Use:   "cdp <fileSHA256>",
	Short: "Test ChromeDP",
	Long:  `Test ChromeDP and get browser version.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("[DEBUG] CMD cdp called")
		chromedpTest()

		return nil
	},
}

func chromedpTest() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var version string
	err := chromedp.Run(ctx,
		chromedp.Navigate("chrome://settings/help"),
		// chromedp.GetVersion(&version)); err != nil { 		log.Fatal(err) 	 } // Thanks for the AI Trip...
		chromedp.Evaluate(`
		(function() {
		const selector = document.querySelector("body > settings-ui").shadowRoot.querySelector("#main").shadowRoot.querySelector("settings-about-page").shadowRoot.querySelector("settings-section:nth-child(8) > div:nth-child(2) > div.flex.cr-padded-text > div.secondary");
		if (!selector) return "selector not found";
		return selector.innerHTML;
			})()
		`, &version),
	)
	if err != nil {
		log.Fatal(err)
	}
	chromedp.Cancel(ctx)
	if version != "" {
		fmt.Println("Chrome version:", version)
	}
}

func init() {
	rootCmd.AddCommand(cdpCmd)
}

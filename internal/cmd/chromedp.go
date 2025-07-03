package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

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
	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true),
	)
	ctx, cancel = chromedp.NewExecAllocator(ctx, options...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(
		ctx,
		chromedp.WithDebugf(log.Printf),
	)
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

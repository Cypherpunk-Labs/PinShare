package psfs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	// 	cid "github.com/ipfs/go-cid"
	// mh "github.com/multiformats/go-multihash"
)

// REF: https://cid.ipfs.tech/

// func getCID(filePath string) (string, error) {
// 	// 1. Open the file
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return "", fmt.Errorf("error opening file: %w", err)
// 	}
// 	defer file.Close()

// 	fi, err := file.Stat()
// 	if err != nil {
// 		return "", fmt.Errorf("error getting file info: %w", err)
// 	}

// 	// 2. Set up the IPLD machinery.
// 	// An in-memory blockstore is used to store the DAG blocks.
// 	bs := blockstore.NewBlockstore(datastore.NewDatastore())
// 	lsys := cidlink.DefaultLinkSystem()
// 	lsys.StorageReadOpener = func(lctx linking.LinkContext, lnk datamodel.Link) (fs.File, error) {
// 		c, ok := lnk.(cid.Cid)
// 		if !ok {
// 			return nil, fmt.Errorf("unexpected link type")
// 		}
// 		blk, err := bs.Get(lctx.Ctx, c)
// 		if err != nil {
// 			return nil, err
// 		}
// 		return &helpers.BlockReadOpener{Block: blk}, nil
// 	}
// 	lsys.StorageWriteOpener = func(lctx linking.LinkContext) (fs.File, datamodel.BlockWriteCommitter, error) {
// 		buf := helpers.NewBlockWriteBuffer()
// 		return buf, buf, nil
// 	}

// 	// 3. Configure the importer parameters
// 	// Using CidV1, sha2-256, and raw leaves to match the command:
// 	// `ipfs add --cid-version=1 --raw-leaves`
// 	prefix, err := cid.PrefixForV1(cid.DagProtobuf, cid.SHA2_256)
// 	if err != nil {
// 		return "", fmt.Errorf("error creating CID prefix: %w", err)
// 	}

// 	params := helpers.DagBuilderParams{
// 		Maxlinks:   helpers.DefaultLinksPerBlock,
// 		RawLeaves:  true, // This corresponds to the --raw-leaves flag
// 		CidBuilder: &prefix,
// 		Dagserv:    &helpers.DagServ{Bstore: bs},
// 	}

// 	db, err := params.New(chunker.NewSizeSplitter(file, chunker.DefaultBlockSize))
// 	if err != nil {
// 		return "", fmt.Errorf("error creating dag builder: %w", err)
// 	}

// 	// 4. Build the DAG
// 	node, err := balanced.Layout(db)
// 	if err != nil {
// 		return "", fmt.Errorf("error laying out DAG: %w", err)
// 	}

// 	// 5. Get the root CID
// 	finalCid := node.Cid()

// 	fmt.Printf("Successfully generated CID: %s\n", finalCid.String())
// 	return finalCid.String(), nil
// }

// func getCID(filePath string) (string, error) {

// 	// // TODO Block01: not correct code on small file nor large
// 	// data, err := ioutil.ReadFile(filePath)
// 	// if err != nil {
// 	// 	fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
// 	// 	os.Exit(1)
// 	// }

// 	// fileNode := unixfs.NewFSNode(unixfs.TFile)
// 	// fileNode.AddBlockSize(uint64(len(data)))
// 	// fileNode.SetData(data)

// 	// serialized, err := fileNode.GetBytes()
// 	// if err != nil {
// 	// 	return "", fmt.Errorf("error serializing UnixFS node: %v", err)
// 	// }

// 	// hash, err := mh.Sum(serialized, mh.SHA2_256, -1)
// 	// if err != nil {
// 	// 	return "", fmt.Errorf("error creating hash: %v", err)
// 	// }

// 	// // cid := cid.NewCidV1(cid.DagProtobuf, hash)
// 	// cid := cid.NewCidV1(cid.Raw, hash)
// 	// fmt.Println(hash)
// 	// // TODO Block01:

// 	// TODO Block 02: seems to work for smaller file, but not larger
// 	sha256, _ := GetSHA256(filePath)
// 	hxhash, _ := hex.DecodeString("1220" + sha256)
// 	cid := cid.NewCidV1(cid.Raw, mh.Multihash(hxhash))
// 	// TODO Block 02

// 	fmt.Printf("CID: %s\n", cid.String())
// 	return cid.String(), nil
// }

// func validateCID(cidString string) (bool, error) {
// 	cidObj, err := cid.Decode(cidString)
// 	if err != nil {
// 		return false, err
// 	}
// 	fmt.Print(cidObj.Hash()) // 122064936ff52a67ed4c029521fd3fbaa1c66a3689f6437af929e6cd7c9897da8112
// 	return true, nil
// }

func GetSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	return hashString, nil
}

// container error
// 2025/07/02 10:53:39 page load error net::ERR_CONNECTION_TIMED_OUT
func GetVirusTotalVerdictByHash(hash string) (bool, error) {
	// safe == true
	// unsafe == false
	baseurl := "https://www.virustotal.com"
	uri := "/gui/file/"
	url := baseurl + uri + hash

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"),
	)
	allocctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	// var screenshotBuffer []byte
	var htmlContent string
	cdpctx, cancel := chromedp.NewContext(allocctx)
	defer cancel()

	err := chromedp.Run(cdpctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),

		chromedp.Evaluate(`
			(function() {
				// First shadow root (file-view)
				const fileView = document.querySelector("file-view");
				if (!fileView) return "file-view not found";

				// Second shadow root (vt-ui-main-generic-report)
				const report = fileView.shadowRoot.querySelector("vt-ui-main-generic-report");
				if (!report) return "vt-ui-main-generic-report not found";

				// Navigate to vt-ioc-score-widget
				const scoreWidget = report.shadowRoot.querySelector("div > div:nth-child(1) > div:nth-child(1) > vt-ioc-score-widget");
				if (!scoreWidget) return "vt-ioc-score-widget not found";

				// Third shadow root (vt-ioc-score-widget)
				const innerWidget = scoreWidget.shadowRoot.querySelector("div > vt-ioc-score-widget-detections-chart");
				if (!innerWidget) return "vt-ioc-score-widget-detections-chart not found";

				// Fourth shadow root (vt-ioc-score-widget-detections-chart)
				const chart = innerWidget.shadowRoot.querySelector("div > div > div:nth-child(1)");
				if (!chart) return "Target div not found";

				return chart.innerHTML;
			})()
		`, &htmlContent),
	)
	if err != nil {
		log.Fatal(err)
	}
	chromedp.Cancel(cdpctx)

	// err = os.WriteFile("screenshot.png", screenshotBuffer, 00644)
	// if err != nil {
	// 	log.Fatal("Error:", err)
	// }

	// expecting " <!--?lit$045892178$-->0 " or " <!--?lit$521644774$-->65 "
	if htmlContent != "" {
		split := strings.Split(htmlContent, ">")
		if len(split) > 1 {
			i, err := strconv.Atoi(strings.TrimSpace(split[1]))
			if err != nil {
				return false, err // TODO: test this is hit when no report and record the error to test for.
			}
			if i == 0 {
				return true, nil
			}
		} else {
			// "file-view not found"
			// TODO: Log error is no report exists
			return false, nil
		}
	}
	return false, nil
	// BUG: After some hours some other response is received, somehow leading to a true response that accepts file into metadata and filesystem on both sides.
}

func SendFileToVirusTotal(inputfilepath string) (bool, error) {
	baseurl := "https://www.virustotal.com"
	uri := "/gui/home/upload"
	url := baseurl + uri

	absPath, err := filepath.Abs(inputfilepath)
	if err != nil {

	}

	options := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", false),
	)
	allocctx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	defer cancel()

	cdpctx, cancel := chromedp.NewContext(allocctx)
	defer cancel()

	var dialogMessage string
	chromedp.ListenTarget(cdpctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *page.EventFileChooserOpened:
			go func(backendNodeID cdp.BackendNodeID) {
				if err := chromedp.Run(cdpctx,
					dom.SetFileInputFiles([]string{absPath}).
						WithBackendNodeID(backendNodeID),
				); err != nil {
					log.Fatal(err)
				}
			}(ev.BackendNodeID)
		}
	})
	if dialogMessage == "" {
	}

	var ids []cdp.NodeID
	var htmlContent string

	// selector1 := `document.querySelector('home-view').shadowRoot.querySelector('vt-ui-main-upload-form').shadowRoot.querySelector('#infoIcon')`
	selector2 := `document.querySelector("#view-container > home-view").shadowRoot.querySelector("#uploadForm").shadowRoot.querySelector("#infoIcon")`
	err = chromedp.Run(cdpctx,
		page.SetInterceptFileChooserDialog(true),
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),

		chromedp.NodeIDs(selector2, &ids, chromedp.ByJSPath),
		chromedp.ActionFunc(func(cdpctx context.Context) error {
			if len(ids) < 1 {
				return fmt.Errorf("[ERROR] selector %q did not return any nodes", ids)
			}
			err := dom.Focus().WithNodeID(ids[0]).Do(cdpctx)
			if err != nil {
				return err
			}
			chromedp.KeyEvent(kb.Enter).Do(cdpctx)
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = chromedp.Run(cdpctx,
		chromedp.WaitVisible(`document.querySelector('file-view')`, chromedp.ByJSPath),
		chromedp.Evaluate(`
			(function() {
				// First shadow root (file-view)
				const fileView = document.querySelector("file-view");
				if (!fileView) return "file-view not found";

				// Second shadow root (vt-ui-main-generic-report)
				const report = fileView.shadowRoot.querySelector("vt-ui-main-generic-report");
				if (!report) return "vt-ui-main-generic-report not found";

				// Navigate to vt-ioc-score-widget
				const scoreWidget = report.shadowRoot.querySelector("div > div:nth-child(1) > div:nth-child(1) > vt-ioc-score-widget");
				if (!scoreWidget) return "vt-ioc-score-widget not found";

				// Third shadow root (vt-ioc-score-widget)
				const innerWidget = scoreWidget.shadowRoot.querySelector("div > vt-ioc-score-widget-detections-chart");
				if (!innerWidget) return "vt-ioc-score-widget-detections-chart not found";

				// Fourth shadow root (vt-ioc-score-widget-detections-chart)
				const chart = innerWidget.shadowRoot.querySelector("div > div > div:nth-child(1)");
				if (!chart) return "Target div not found";

				return chart.innerHTML;
			})()
		`, &htmlContent),
	)
	chromedp.Cancel(cdpctx)

	if htmlContent != "" {
		split := strings.Split(htmlContent, ">")
		if len(split) > 1 {
			i, err := strconv.Atoi(strings.TrimSpace(split[1]))
			if err != nil {
				return false, err // TODO: test this is hit when no report and record the error to test for.
			}
			if i == 0 {
				return true, nil
			}
		} else {
			// "file-view not found"
			// TODO: Log error is no report exists
			return false, nil
		}
	}
	return false, nil
}

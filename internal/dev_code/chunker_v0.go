package localfs

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"log"

// 	"github.com/ipfs/go-cid"
// 	chunker "github.com/ipfs/go-ipfs-chunker"
// 	"github.com/ipld/go-ipld-prime"
// 	"github.com/ipld/go-ipld-prime/codec/raw"
// 	"github.com/ipld/go-ipld-prime/datamodel"
// 	"github.com/ipld/go-ipld-prime/linking"
// 	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
// 	"github.com/ipld/go-ipld-prime/node/basicnode"
// 	"github.com/multiformats/go-multihash"
// )

// // createLeafNode creates an IPLD node from a data chunk and returns its CID
// func createLeafNode(data []byte) (datamodel.Node, cid.Cid, error) {
// 	// Create a basic IPLD node from the data
// 	node := basicnode.NewBytes(data)

// 	// Set up a CID prefix
// 	prefix := cid.Prefix{
// 		Version:  1,       // CIDv1
// 		Codec:    cid.Raw, // Raw binary data
// 		MhType:   multihash.SHA2_256,
// 		MhLength: -1, // Use default length
// 	}

// 	// Create a link system to compute the CID
// 	ls := cidlink.DefaultLinkSystem()
// 	ls.EncoderChooser = func(_ datamodel.LinkPrototype) (ipld.Encoder, error) {
// 		return raw.Encode, nil // Register raw codec encoder
// 	}
// 	var cidOut cid.Cid
// 	ls.StorageWriteOpener = func(_ linking.LinkContext) (io.Writer, linking.BlockWriteCommitter, error) {
// 		var buf bytes.Buffer
// 		return &buf, func(l datamodel.Link) error {
// 			cidOut = l.(cidlink.Link).Cid
// 			return nil
// 		}, nil
// 	}

// 	// Store the node to compute its CID
// 	lp := cidlink.LinkPrototype{Prefix: prefix}
// 	_, err := ls.Store(linking.LinkContext{}, lp, node)
// 	if err != nil {
// 		return nil, cid.Undef, err
// 	}

// 	return node, cidOut, nil
// }

// func chunker_test() {
// 	// Sample data to chunk
// 	// data := []byte("This is a sample text that will be chunked and stored in IPFS.")
// 	data, _ := ioutil.ReadFile("/Users/mkemp/repos/cypherpunk-labs/test-data-ai-etl/docs/pdf/Nagel IE163 Direct Electrical Production from LENR.pdf")

// 	// Create a chunker with a size of 256 bytes
// 	splitter := chunker.NewSizeSplitter(bytes.NewReader(data), 256)

// 	var cids []cid.Cid
// 	var nodes []datamodel.Node

// 	// Process each chunk
// 	for {
// 		chunk, err := splitter.NextBytes()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			log.Fatalf("Error reading chunk: %v", err)
// 		}

// 		// Create an IPLD node and CID for the chunk
// 		node, chunkCid, err := createLeafNode(chunk)
// 		if err != nil {
// 			log.Fatalf("Error creating leaf node: %v", err)
// 		}

// 		cids = append(cids, chunkCid)
// 		nodes = append(nodes, node)
// 	}

// 	// Create a root node linking to all chunk CIDs (simple list for this example)
// 	rootBuilder := basicnode.Prototype.List.NewBuilder()
// 	rootAssembler, err := rootBuilder.BeginList(int64(len(cids)))
// 	if err != nil {
// 		log.Fatalf("Error creating root node: %v", err)
// 	}

// 	for _, c := range cids {
// 		link := cidlink.Link{Cid: c}
// 		if err := rootAssembler.AssembleValue().AssignLink(link); err != nil {
// 			log.Fatalf("Error assembling link: %v", err)
// 		}
// 	}
// 	if err := rootAssembler.Finish(); err != nil {
// 		log.Fatalf("Error finishing root node: %v", err)
// 	}
// 	rootNode := rootBuilder.Build()

// 	// Compute CID for the root node
// 	_, rootCid, err := createLeafNode([]byte(fmt.Sprintf("%v", rootNode)))
// 	if err != nil {
// 		log.Fatalf("Error creating root CID: %v", err)
// 	}

// 	// Print the results
// 	fmt.Printf("Root CID: %s\n", rootCid.String())
// 	fmt.Println("Chunk CIDs:")
// 	for i, c := range cids {
// 		fmt.Printf("Chunk %d: %s\n", i+1, c.String())
// 	}
// }

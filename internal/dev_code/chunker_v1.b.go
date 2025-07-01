package localfs

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"log"

// 	"github.com/ipfs/go-cid"
// 	chunker "github.com/ipfs/go-ipfs-chunker"
// 	"github.com/ipfs/go-unixfs"
// 	dagpb "github.com/ipld/go-dagpb"
// 	"github.com/ipld/go-ipld-prime"
// 	"github.com/ipld/go-ipld-prime/codec/dagpb"
// 	"github.com/ipld/go-ipld-prime/datamodel"
// 	"github.com/ipld/go-ipld-prime/linking"
// 	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
// 	"github.com/multiformats/go-multihash"
// )

// // createLeafNode creates a DAG-PB node for a data chunk and returns its CID
// func createLeafNode(data []byte) (datamodel.Node, cid.Cid, error) {
// 	// Create a UnixFS data block
// 	fsNode := unixfs.NewFSNode(unixfs.TFile)
// 	fsNode.SetData(data)

// 	// Build the DAG-PB node
// 	nodeBuilder := dagpb.Type.PBNode.NewBuilder()
// 	nodeMap, err := nodeBuilder.BeginMap(2)
// 	if err != nil {
// 		return nil, cid.Undef, err
// 	}
// 	if err := nodeMap.AssembleKey().AssignString("Data"); err != nil {
// 		return nil, cid.Undef, err
// 	}
// 	dataBytes, err := fsNode.GetBytes()
// 	if err != nil {
// 		return nil, cid.Undef, err
// 	}
// 	if err := nodeMap.AssembleValue().AssignBytes(dataBytes); err != nil {
// 		return nil, cid.Undef, err
// 	}
// 	if err := nodeMap.AssembleKey().AssignString("Links"); err != nil {
// 		return nil, cid.Undef, err
// 	}
// 	if _, err := nodeMap.AssembleValue().BeginList(0); err != nil {
// 		return nil, cid.Undef, err
// 	}
// 	if err := nodeMap.Finish(); err != nil {
// 		return nil, cid.Undef, err
// 	}
// 	node := nodeBuilder.Build()

// 	// Set up a CID prefix
// 	prefix := cid.Prefix{
// 		Version:  1,
// 		Codec:    cid.DagProtobuf,
// 		MhType:   multihash.SHA2_256,
// 		MhLength: -1,
// 	}

// 	// Create a link system to compute the CID
// 	ls := cidlink.DefaultLinkSystem()
// 	ls.EncoderChooser = func(_ datamodel.LinkPrototype) (ipld.Encoder, error) {
// 		return dagpb.Encode, nil
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
// 	_, err = ls.Store(linking.LinkContext{}, lp, node)
// 	if err != nil {
// 		return nil, cid.Undef, err
// 	}

// 	return node, cidOut, nil
// }

// func chunk_test() {
// 	// Sample data to chunk
// 	// data := []byte("This is a sample text that will be chunked and stored in IPFS.")
// 	data, _ := ioutil.ReadFile("/Users/mkemp/repos/cypherpunk-labs/test-data-ai-etl/docs/pdf/Nagel IE163 Direct Electrical Production from LENR.pdf")

// 	// Create a chunker with IPFS default size (256 KiB)
// 	splitter := chunker.NewSizeSplitter(bytes.NewReader(data), 262144)

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

// 		// Create a DAG-PB node and CID for the chunk
// 		node, chunkCid, err := createLeafNode(chunk)
// 		if err != nil {
// 			log.Fatalf("Error creating leaf node: %v", err)
// 		}

// 		cids = append(cids, chunkCid)
// 		nodes = append(nodes, node)
// 	}

// 	// Create a UnixFS root node
// 	fsNode := unixfs.NewFSNode(unixfs.TFile)
// 	for _, c := range cids {
// 		fsNode.AddBlockSize(uint64(c.ByteLen()))
// 	}

// 	// Build the root DAG-PB node
// 	rootBuilder := dagpb.Type.PBNode.NewBuilder()
// 	rootMap, err := rootBuilder.BeginMap(2)
// 	if err != nil {
// 		log.Fatalf("Error creating root node: %v", err)
// 	}
// 	if err := rootMap.AssembleKey().AssignString("Data"); err != nil {
// 		log.Fatalf("Error assembling root node: %v", err)
// 	}
// 	dataBytes, err := fsNode.GetBytes()
// 	if err != nil {
// 		log.Fatalf("Error getting root node data: %v", err)
// 	}
// 	if err := rootMap.AssembleValue().AssignBytes(dataBytes); err != nil {
// 		log.Fatalf("Error assembling root node: %v", err)
// 	}
// 	if err := rootMap.AssembleKey().AssignString("Links"); err != nil {
// 		log.Fatalf("Error assembling root node: %v", err)
// 	}
// 	linksList, err := rootMap.AssembleValue().BeginList(int64(len(cids)))
// 	if err != nil {
// 		log.Fatalf("Error creating links list: %v", err)
// 	}
// 	for _, c := range cids {
// 		linkMap, err := linksList.AssembleValue().BeginMap(3)
// 		if err != nil {
// 			log.Fatalf("Error creating link: %v", err)
// 		}
// 		if err := linkMap.AssembleKey().AssignString("Hash"); err != nil {
// 			log.Fatalf("Error assembling link: %v", err)
// 		}
// 		if err := linkMap.AssembleValue().AssignLink(cidlink.Link{Cid: c}); err != nil {
// 			log.Fatalf("Error assigning link: %v", err)
// 		}
// 		if err := linkMap.AssembleKey().AssignString("Name"); err != nil {
// 			log.Fatalf("Error assembling link: %v", err)
// 		}
// 		if err := linkMap.AssembleValue().AssignString(""); err != nil {
// 			log.Fatalf("Error assigning link name: %v", err)
// 		}
// 		if err := linkMap.AssembleKey().AssignString("TSize"); err != nil {
// 			log.Fatalf("Error assembling link: %v", err)
// 		}
// 		if err := linkMap.AssembleValue().AssignInt(int64(c.ByteLen())); err != nil {
// 			log.Fatalf("Error assigning link size: %v", err)
// 		}
// 		if err := linkMap.Finish(); err != nil {
// 			log.Fatalf("Error finishing link: %v", err)
// 		}
// 	}
// 	if err := linksList.Finish(); err != nil {
// 		log.Fatalf("Error finishing links: %v", err)
// 	}
// 	if err := rootMap.Finish(); err != nil {
// 		log.Fatalf("Error finishing root node: %v", err)
// 	}
// 	rootNode := rootBuilder.Build()

// 	// Compute CID for the root node
// 	_, rootCid, err := createLeafNode(dataBytes)
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

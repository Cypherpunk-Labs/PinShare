package localfs

// import (
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"io"
// 	"os"

// 	"github.com/ipfs/go-blockservice"
// 	cid "github.com/ipfs/go-cid"
// 	"github.com/ipfs/go-datastore"
// 	dssync "github.com/ipfs/go-datastore/sync"
// 	multihash "github.com/multiformats/go-multihash/core"

// 	blockstore "github.com/ipfs/go-ipfs-blockstore"
// 	// blockstore "github.com/ipfs/boxo/blockservice"
// 	chunker "github.com/ipfs/go-ipfs-chunker"

// 	"github.com/ipfs/go-merkledag"
// 	// "github.com/ipfs/boxo/ipld/merkledag"
// )

// func chunker_test() {
// 	file, _ := os.ReadFile("../../test/test02/gt256kb.txt")
// 	// file, _ := os.ReadFile("../../test/test02/small.txt")

// 	data := bytes.NewReader(file)

// 	// Example data to mimic a file
// 	// data := bytes.NewReader([]byte("This is some example data to be chunked and added to IPFS."))

// 	// Build the UnixFS DAG
// 	ctx := context.Background()
// 	cid, err := buildUnixFSDAG(ctx, data)
// 	if err != nil {
// 		fmt.Printf("Error building DAG: %v\n", err)
// 		return
// 	}

// 	// Print the CID
// 	fmt.Printf("Generated CID: %s\n", cid.String())
// }

// // buildUnixFSDAG constructs a UnixFS DAG from the chunked data
// func buildUnixFSDAG(ctx context.Context, data *bytes.Reader) (cid.Cid, error) {
// 	// Create an in-memory datastore
// 	ds := dssync.MutexWrap(datastore.NewMapDatastore())

// 	bs := blockservice.New(blockstore.NewBlockstore(ds), nil) // Use default offline exchange
// 	dagService := merkledag.NewDAGService(bs)

// 	// Create a fixed-size chunker (256KB, default for ipfs add)
// 	splitter := chunker.NewSizeSplitter(data, 256*1024)

// 	// Initialize variables for the DAG
// 	var totalSize uint64
// 	// var links []*format.Link

// 	builder := cid.NewPrefixV1(cid.DagProtobuf, multihash.SHA2_256)
// 	rootb := &merkledag.ProtoNode{}
// 	rootb.SetCidBuilder(builder)

// 	// Process each chunk
// 	for {
// 		chunk, err := splitter.NextBytes()
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			return cid.Undef, fmt.Errorf("error reading chunk: %w", err)
// 		}

// 		totalSize += uint64(len(chunk))

// 		// Create a raw node for the chunk
// 		leafNode := merkledag.NewRawNode(chunk)

// 		// Add the node to the DAG
// 		if err := dagService.Add(ctx, leafNode); err != nil {
// 			return cid.Undef, fmt.Errorf("error adding node to DAG: %w", err)
// 		}

// 		// lnc := leafNode.Cid()
// 		// rootb := &merkledag.ProtoNode{}
// 		rootb.AddNodeLink(leafNode.Cid().String(), leafNode)
// 		// Create a link to the chunk
// 		// link, err := merkledag.NewLink(leafNode, leafNode.Cid().String())
// 		// link := rootb.Links() // merkledag.GetLinks(leafNode, lnc)
// 		// if err != nil {
// 		// 	return cid.Undef, fmt.Errorf("error creating link: %w", err)
// 		// }
// 		// links = append(links, link[0])
// 	}

// 	// Create the root UnixFS file node
// 	// fsNode := new(unixfs.FSNode) // unixfs.NewFSNode(unixfs.TFile)
// 	// // fsNode.SetFileSize(totalSize)
// 	// fsNode.BlockSize(int(totalSize))

// 	// Build the DAG with the UnixFS builder
// 	// b := unixfs.NewBuilder()
// 	// b := unixfs.NewBuilder() // dunno how to fix
// 	// b.SetLinks(links)
// 	// b.SetNode(fsNode)

// 	// // ipfs.
// 	// 	rootnode, _ := cid.V1Builder(ctx)
// 	// rootnode.Cid()

// 	// rootNode := b.Build(ctx, dagService)
// 	// if err != nil {
// 	// 	return cid.Undef, fmt.Errorf("error building root node: %w", err)
// 	// }

// 	println(rootb.Cid().String())

// 	return cid.Cid{}, nil
// 	//return rootNode.Cid(), nil
// }

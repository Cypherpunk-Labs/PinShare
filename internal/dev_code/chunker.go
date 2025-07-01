package localfs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ipfs/boxo/blockservice"
	cid "github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	offline "github.com/ipfs/go-ipfs-exchange-offline"

	// "github.com/ipfs/go-unixfs/importer/balanced"
	"github.com/ipfs/boxo/ipld/unixfs/importer/balanced"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"

	blockstore "github.com/ipfs/boxo/blockstore"
	// blockstore "github.com/ipfs/boxo/blockservice"
	chunker "github.com/ipfs/boxo/chunker"

	"github.com/ipfs/boxo/ipld/merkledag"
	// "github.com/ipfs/boxo/ipld/merkledag"
	uih "github.com/ipfs/boxo/ipld/unixfs/importer/helpers"
)

func Cid(filename string) string {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	fileReader := bytes.NewReader(fileData)
	ds := dssync.MutexWrap(datastore.NewNullDatastore())
	bs := blockstore.NewBlockstore(ds)
	bs = blockstore.NewIdStore(bs)
	bsrv := blockservice.New(bs, offline.Exchange(bs))
	dsrv := merkledag.NewDAGService(bsrv)
	ufsImportParams := uih.DagBuilderParams{
		Maxlinks:  uih.DefaultLinksPerBlock, // Default max of 174 links per block
		RawLeaves: true,
		CidBuilder: cid.V1Builder{ // Use CIDv1 for all links
			Codec:    uint64(multicodec.Raw),
			MhType:   uint64(multicodec.Sha2_256), //SHA2-256
			MhLength: -1,
		},
		Dagserv: dsrv,
		NoCopy:  false,
	}
	ufsBuilder, err := ufsImportParams.New(chunker.NewSizeSplitter(fileReader, chunker.DefaultBlockSize)) // 256KiB chunks
	if err != nil {
		return cid.Undef.String()
	}
	nd, err := balanced.Layout(ufsBuilder)
	if err != nil {
		return cid.Undef.String()
	}
	fmt.Println(nd.Cid().String())
	return nd.Cid().String()
}

func chunker_test(inputFile string) string {
	// This method does not match IPFS add cmd
	file, _ := os.ReadFile(inputFile)
	// file, _ := os.ReadFile("../../test/test02/gt256kb.txt")
	// file, _ := os.ReadFile("../../test/test02/small.txt")

	data := bytes.NewReader(file)

	// Example data to mimic a file
	// data := bytes.NewReader([]byte("This is some example data to be chunked and added to IPFS."))

	// Build the UnixFS DAG
	ctx := context.Background()
	cid, err := buildUnixFSDAG(ctx, data)
	if err != nil {
		fmt.Printf("Error building DAG: %v\n", err)
		return ""
	}

	// Print the CID
	// fmt.Printf("Generated CID: %s\n", cid.String())
	fmt.Printf("2:%s\n", cid.String())
	return cid.String()
}

// buildUnixFSDAG constructs a UnixFS DAG from the chunked data
func buildUnixFSDAG(ctx context.Context, data *bytes.Reader) (cid.Cid, error) {
	// Create an in-memory datastore
	ds := dssync.MutexWrap(datastore.NewMapDatastore())

	bs := blockservice.New(blockstore.NewBlockstore(ds), nil) // Use default offline exchange
	dagService := merkledag.NewDAGService(bs)

	// Create a fixed-size chunker (256KB, default for ipfs add)
	splitter := chunker.NewSizeSplitter(data, 256*1024)

	// Initialize variables for the DAG
	var totalSize uint64
	// var links []*format.Link

	builder := cid.NewPrefixV1(cid.DagProtobuf, multihash.SHA2_256)
	rootb := &merkledag.ProtoNode{}
	rootb.SetCidBuilder(builder)

	// Process each chunk
	for {
		chunk, err := splitter.NextBytes()
		if err == io.EOF {
			break
		}
		if err != nil {
			return cid.Undef, fmt.Errorf("error reading chunk: %w", err)
		}

		totalSize += uint64(len(chunk))

		// Create a raw node for the chunk
		leafNode := merkledag.NewRawNode(chunk)

		// Add the node to the DAG
		if err := dagService.Add(ctx, leafNode); err != nil {
			return cid.Undef, fmt.Errorf("error adding node to DAG: %w", err)
		}

		rootb.AddNodeLink(leafNode.Cid().String(), leafNode)
	}

	//println(rootb.Cid().String())

	// return cid.Cid{}, nil
	return rootb.Cid(), nil
}

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"pinshare/internal/p2p"
	"pinshare/internal/store"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server implements the ServerInterface.
type Server struct{}

// NewServer creates a new server.
func NewServer() *Server {
	return &Server{}
}

// writeError is a helper for sending a JSON error response.
func writeError(w http.ResponseWriter, code int, message string) {
	errResp := Error{
		Code:    func() *int32 { c := int32(code); return &c }(),
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(errResp)
}

// ListAllFiles handles GET /files
func (s *Server) ListAllFiles(w http.ResponseWriter, r *http.Request) {
	allStoreFiles := store.GlobalStore.GetAllFiles()
	apiFiles := make([]BaseMetadata, len(allStoreFiles))
	// TODO: should we filter out the bansets if so we need to adjust our length. or use add instead of make

	for i, f := range allStoreFiles {
		// Create copies of the time values to take their addresses.
		lastUpdated := f.LastUpdated
		addedAt := f.AddedAt
		apiFiles[i] = BaseMetadata{
			FileSHA256:  f.FileSHA256,
			IpfsCID:     f.IPFSCID,
			FileType:    f.FileType,
			LastUpdated: &lastUpdated,
			AddedAt:     &addedAt,
			// TODO: add remaining fields
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiFiles)
}

// GetFileBySHA256 handles GET /files/{fileSHA256}
func (s *Server) GetFileBySHA256(w http.ResponseWriter, r *http.Request, fileSHA256 string) {
	file, found := store.GlobalStore.GetFile(fileSHA256)
	if !found {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	lastUpdated := file.LastUpdated
	addedAt := file.AddedAt
	apiFile := BaseMetadata{
		FileSHA256:  file.FileSHA256,
		IpfsCID:     file.IPFSCID,
		FileType:    file.FileType,
		LastUpdated: &lastUpdated,
		AddedAt:     &addedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiFile)
}

// AddOrUpdateFile handles PUT /files/{fileSHA256}
func (s *Server) AddOrUpdateFile(w http.ResponseWriter, r *http.Request, fileSHA256 string) {
	var body WritableMetadata
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Check if the file already exists to determine status code (200 vs 201).
	_, exists := store.GlobalStore.GetFile(fileSHA256)

	meta := store.BaseMetadata{
		FileSHA256: fileSHA256,
		IPFSCID:    body.IpfsCID,
		FileType:   body.FileType,
	}

	if err := store.GlobalStore.AddFile(meta); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to add or update file")
		return
	}

	// Retrieve the updated metadata to return it in the response.
	updatedFile, _ := store.GlobalStore.GetFile(fileSHA256)
	lastUpdated := updatedFile.LastUpdated
	addedAt := updatedFile.AddedAt
	apiFile := BaseMetadata{
		FileSHA256:  updatedFile.FileSHA256,
		IpfsCID:     updatedFile.IPFSCID,
		FileType:    updatedFile.FileType,
		LastUpdated: &lastUpdated,
		AddedAt:     &addedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	if exists {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	_ = json.NewEncoder(w).Encode(apiFile)
}

func (s *Server) AddTag(w http.ResponseWriter, r *http.Request, fileSHA256 string) {
	// TODO:
}

func (s *Server) RemoveTag(w http.ResponseWriter, r *http.Request, fileSHA256 string, tagName string) {
	// TODO:
}

func (s *Server) VoteForRemoval(w http.ResponseWriter, r *http.Request, fileSHA256 string) {
	// TODO:
}

func (s *Server) ListP2PPeers(w http.ResponseWriter, r *http.Request) {
	node := GetNode()
	if node == nil {
		writeError(w, http.StatusInternalServerError, "P2P node not initialized")
		return
	}

	peers := (*node).Network().Peers()
	peerIDs := make([]string, len(peers))
	for i, p := range peers {
		peerIDs[i] = p.String()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(peerIDs)
}

func (s *Server) ConnectToPeer(w http.ResponseWriter, r *http.Request) {
	var body ConnectToPeerJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	node := GetNode()
	if node == nil {
		writeError(w, http.StatusInternalServerError, "P2P node not initialized")
		return
	}

	// Parse the multiaddress from the request body.
	peerMA, err := multiaddr.NewMultiaddr(body.Multiaddr)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid multiaddress: %v", err))
		return
	}

	// Extract peer info from the multiaddress.
	peerInfo, err := peer.AddrInfoFromP2pAddr(peerMA)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("could not get peer info from multiaddress: %v", err))
		return
	}

	// Run connection attempt in the background to not block the API response.
	go func() {
		log.Printf("[API-INFO] Attempting to connect to peer %s", peerInfo.ID.String())
		if err := (*node).Connect(context.Background(), *peerInfo); err != nil {
			log.Printf("[API-WARN] Failed to connect to peer %s: %v", peerInfo.ID.String(), err)
		} else {
			log.Printf("[API-INFO] Successfully connected to peer %s", peerInfo.ID.String())
		}
	}()

	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte("Connection attempt initiated"))
}

func (s *Server) SendDirectMessage(w http.ResponseWriter, r *http.Request, peerID string) {
	var body SendDirectMessageJSONBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	peerIDParsed, err := p2p.ParsePeerID(peerID)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid peer ID: %v", err))
		return
	}

	err = p2p.DirectMessagePeer(context.Background(), *GetNode(), peerIDParsed, []byte(body.Message))
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to send direct message: %v", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message sent successfully"))
}

func (s *Server) GetP2PStatus(w http.ResponseWriter, r *http.Request) {
	node := GetNode()
	if node == nil {
		writeError(w, http.StatusInternalServerError, "P2P node not initialized")
		return
	}

	addrs := make([]string, 0, len((*node).Addrs()))
	for _, addr := range (*node).Addrs() {
		addrs = append(addrs, addr.String())
	}

	status := P2PStatus{
		Id:             (*node).ID().String(),
		Addresses:      addrs,
		ConnectedPeers: len((*node).Network().Peers()),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(status)
}

var p2pNodeInstance *host.Host

// SetP2PManager allows main to set the global PubSubManager instance
func SetNode(node *host.Host) {
	p2pNodeInstance = node
}

func GetNode() *host.Host {
	return p2pNodeInstance
}

func Start(ctx context.Context, node host.Host) {
	SetNode(&node)
	server := NewServer()

	// get an `http.Handler` that we can use
	apiHandler := Handler(server)

	// Create a new ServeMux to combine the API handler and metrics handler
	mux := http.NewServeMux()
	mux.Handle("/", apiHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// Check if port 8080 is in use. If so, increment until an open port is found.
	var port int = 9090
	for {
		addr := fmt.Sprintf("0.0.0.0:%d", port)
		conn, err := net.Listen("tcp", addr)
		if err != nil {
			fmt.Printf("[INFO] Port %d is in use, trying next...\n", port)
			port++
			continue
		}
		conn.Close()
		break
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	s := &http.Server{
		Handler: mux,
		Addr:    addr,
	}

	log.Printf("[INFO] Starting API server on %s", addr)
	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}

package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"pinshare/internal/store"
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

func Start() {
	server := NewServer()

	// get an `http.Handler` that we can use
	h := Handler(server)

	// generate random number between 01 and 99
	// TODO: This is a temporary solution for testing purposes.
	// In a real application, the port should be configurable.
	min := 1
	max := 99
	randomNumber := rand.Intn(max-min+1) + min
	port := fmt.Sprintf("90%02d", randomNumber)

	// In a real app, you'd likely get this from config.
	// Matching the port from the OpenAPI spec example.
	addr := "0.0.0.0:" + port
	s := &http.Server{
		Handler: h,
		Addr:    addr,
	}

	log.Printf("[INFO] Starting API server on %s", addr)
	// And we serve HTTP until the world ends.
	log.Fatal(s.ListenAndServe())
}

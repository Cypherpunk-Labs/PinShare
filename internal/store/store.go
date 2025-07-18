package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// BaseMetadata holds all metadata for a single file.
// In a full CRDT system, fields like Tags, ModerationVotes, CommunityLabels
// would be specific CRDT types (e.g., OR-Set, PN-Counter).
type BaseMetadata struct {
	FileSHA256      string          `json:"fileSHA256"` // Primary key
	IPFSCID         string          `json:"ipfsCID"`
	FileType        string          `json:"fileType"`
	LastUpdated     time.Time       `json:"lastUpdated"` // Timestamp for LWW or general record update
	AddedAt         time.Time       `json:"addedAt"`
	ModerationVotes int             `json:"moderationVotes"` // Simulating a PN-Counter
	Tags            map[string]bool `json:"tags"`            // Simulating an OR-Set; bool true if tag exists
}

// // TODO: So we will have very specific labels/tags for the content 1-1, and other tags that will label content with 1-many.
// type TagMetadata struct {
// 	IPFSCID string `json:"ipfsCID"`
// 	// TODO: Decide if to use Key Value pairs for tags, and how to keep our tags immutable
// 	Tag string `json:"tag"` // TODO: I want to enable not just a single tag but mutiple tags in an immutable way.
// 	// Title           string          `json:"title"`           // LWW-Register (Last-Write-Wins)
// 	// Author          string          `json:"author"`          // LWW-Register
// 	// FileName        string          `json:"fileName"`        // LWW-Register
// 	// Date            string          `json:"date"`            // LWW-Register (consider time.Time if strict parsing needed)
// 	// ScientificField string          `json:"scientificField"` // LWW-Register
// 	// Tags            map[string]bool `json:"tags"`            // Simulating an OR-Set; bool true if tag exists
// 	ModerationVotes int `json:"moderationVotes"` // Simulating a PN-Counter // TODO: I want to allow users to vote on each Tag only once but also allow them to change at a later time.
// 	// CommunityLabels map[string]int  `json:"communityLabels"` // Simulating a map of Label -> PN-Counter (vote count)
// 	AddedByNodeID string `json:"a"`
// }

// MetadataStore holds all file metadata in memory.
type MetadataStore struct {
	mu       sync.RWMutex
	Files    map[string]*BaseMetadata // Keyed by FileSHA256
	dataFile string
}

// GlobalStore is a global instance for CLI convenience.
// In a real API, avoid global instances.
var GlobalStore = NewMetadataStore()

// NewMetadataStore creates a new empty metadata store.
func NewMetadataStore() *MetadataStore {
	return &MetadataStore{
		Files: make(map[string]*BaseMetadata),
	}
}

// Save persists the current state of the store to a JSON file.
func (s *MetadataStore) Save(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy for marshalling to avoid holding the lock during lengthy I/O
	// and to ensure a consistent snapshot.
	filesCopy := make(map[string]*BaseMetadata)
	for k, v := range s.Files {
		// Deep copy might be needed if FileMetadata contains pointers/slices that could be modified elsewhere
		// For this structure, a shallow copy of the map entries is okay as FileMetadata itself is a struct.
		copiedMeta := *v
		filesCopy[k] = &copiedMeta
	}

	data, err := json.MarshalIndent(filesCopy, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	s.dataFile = filePath // Store the path for future saves if not already set
	fmt.Printf("[DEBUG] Metadata saved to %s\n", filePath)
	return nil
}

// Load populates the store from a JSON file.
func (s *MetadataStore) Load(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File not found is not an error for loading, just means empty store initially
			s.Files = make(map[string]*BaseMetadata)
			s.dataFile = filePath
			return nil
		}
		return fmt.Errorf("failed to read metadata file: %w", err)
	}

	if len(data) == 0 { // Handle empty file case
		s.Files = make(map[string]*BaseMetadata)
		s.dataFile = filePath
		return nil
	}

	err = json.Unmarshal(data, &s.Files)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %w", err)
	}
	// Ensure map is not nil if JSON was "null" or empty
	if s.Files == nil {
		s.Files = make(map[string]*BaseMetadata)
	}
	s.dataFile = filePath
	fmt.Printf("[DEBUG] Metadata loaded from %s\n", filePath)
	return nil
}

// AddFile adds or updates metadata for a file.
// For LWW fields, this simple implementation overwrites. A real LWW CRDT
// would compare timestamps before updating.
func (s *MetadataStore) AddFile(meta BaseMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if meta.FileSHA256 == "" {
		return fmt.Errorf("FileSHA256 cannot be empty")
	}

	now := time.Now().UTC() // TODO should this be with .UTC() also
	existing, exists := s.Files[meta.FileSHA256]
	if !exists {
		meta.AddedAt = now
		// if meta.Tags == nil {
		// 	meta.Tags = make(map[string]bool)
		// }
		// if meta.CommunityLabels == nil {
		// 	meta.CommunityLabels = make(map[string]int)
		// }
	} else {
		// Preserve creation date and potentially merge some fields if needed
		meta.AddedAt = existing.AddedAt
		// For LWW fields, we'd check timestamps here.
		// For POC, we just overwrite, but merge tags.
		// if meta.Tags == nil && existing.Tags != nil { // If new meta has no tags, keep old ones
		// 	meta.Tags = existing.Tags
		// } else if meta.Tags != nil && existing.Tags != nil { // Both have tags, merge them
		// 	for tag, present := range existing.Tags {
		// 		if present { // if tag was present in old
		// 			if _, newHas := meta.Tags[tag]; !newHas { // and not in new
		// 				meta.Tags[tag] = true // add it to new
		// 			}
		// 		}
		// 	}
		// } else if meta.Tags == nil && existing.Tags == nil { // Both nil, initialize
		// 	meta.Tags = make(map[string]bool)
		// }
		// Similar logic for CommunityLabels if needed
		// if meta.CommunityLabels == nil && existing.CommunityLabels != nil {
		// 	meta.CommunityLabels = existing.CommunityLabels
		// } else if meta.CommunityLabels == nil && existing.CommunityLabels == nil {
		// 	meta.CommunityLabels = make(map[string]int)
		// }
	}
	meta.LastUpdated = now
	s.Files[meta.FileSHA256] = &meta
	// No direct saveInternal() call here; commands will call Save explicitly.
	// This makes store methods more unit-testable without file I/O.
	return nil
}

// ApplyGossipUpdate is called when metadata is received from PubSub.
// It implements a simple LWW: if remoteMeta is newer, it's applied.
// Returns true if an update was made, false otherwise.
func (s *MetadataStore) ApplyGossipUpdate(remoteMeta BaseMetadata) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if remoteMeta.FileSHA256 == "" {
		return false, fmt.Errorf("received gossip metadata with empty FileSHA256")
	}

	// Ensure remoteMeta timestamps are UTC for comparison
	remoteMeta.LastUpdated = remoteMeta.LastUpdated
	remoteMeta.AddedAt = remoteMeta.AddedAt

	existing, exists := s.Files[remoteMeta.FileSHA256]
	if !exists || remoteMeta.LastUpdated.After(existing.LastUpdated) {
		// If new, or remote is strictly newer, apply it.
		// We directly store remoteMeta, preserving its LastUpdated timestamp.
		// if remoteMeta.Tags == nil {
		// 	remoteMeta.Tags = make(map[string]bool)
		// }
		// if remoteMeta.CommunityLabels == nil {
		// 	remoteMeta.CommunityLabels = make(map[string]int)
		// }

		s.Files[remoteMeta.FileSHA256] = &remoteMeta
		fmt.Printf("[INFO] Applied gossiped update for %s (Remote LastUpdated: %s)\n", remoteMeta.FileSHA256, remoteMeta.LastUpdated)
		return true, nil
	}

	// fmt.Printf("[DEBUG] Ignored stale gossip update for %s (Remote: %s, Local: %s)\n", remoteMeta.FileSHA256, remoteMeta.LastUpdated, existing.LastUpdated)
	return false, nil
}

// GetFile retrieves metadata for a specific file.
func (s *MetadataStore) GetFile(sha256 string) (*BaseMetadata, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	meta, exists := s.Files[sha256]
	if !exists {
		return nil, false
	}
	// Return a copy to prevent modification of the stored pointer's content from outside the lock
	metaCopy := *meta
	return &metaCopy, true
}

// GetAllFiles retrieves all metadata entries.
func (s *MetadataStore) GetAllFiles() []*BaseMetadata {
	s.mu.RLock()
	defer s.mu.RUnlock()
	all := make([]*BaseMetadata, 0, len(s.Files))
	for _, meta := range s.Files {
		metaCopy := *meta // Return copies
		all = append(all, &metaCopy)
	}
	return all
}

// // TODO: I think this is funtionally equivalent to GetAllFiles() so will change our periodic publisher to use that instead.
// func (s *MetadataStore) GetAllMetadata() []FileMetadata {
//     s.mu.RLock() // Assuming appropriate locking
//     defer s.mu.RUnlock()
//     all := make([]FileMetadata, 0, len(s.Metadata))
//     for _, meta := range s.Metadata {
//         all = append(all, meta)
//     }
//     return all
// }

// AddTag adds a tag to a file. This simulates an OR-Set 'add' operation.
func (s *MetadataStore) AddTag(sha256, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, exists := s.Files[sha256]
	if !exists {
		return fmt.Errorf("file with SHA256 %s not found", sha256)
	}
	if meta.Tags == nil {
		meta.Tags = make(map[string]bool)
	}
	meta.Tags[tag] = true               // In a real OR-Set, this would involve unique IDs per add.
	meta.LastUpdated = time.Now().UTC() // TODO should this be with .UTC() also
	return nil                          // Commands handle saving
}

// RemoveTag removes a tag from a file. This simulates an OR-Set 'remove' operation.
func (s *MetadataStore) RemoveTag(sha256, tag string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, exists := s.Files[sha256]
	if !exists {
		return fmt.Errorf("file with SHA256 %s not found", sha256)
	}
	if meta.Tags == nil {
		// No tags to remove from
		return nil
	}
	delete(meta.Tags, tag)              // In a real OR-Set, this adds a tombstone.
	meta.LastUpdated = time.Now().UTC() // TODO should this be with .UTC() also
	return nil                          // Commands handle saving
}

// VoteForRemoval increments the moderation vote count for a file. Simulates PN-Counter increment.
func (s *MetadataStore) VoteForRemoval(sha256 string, increment bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	meta, exists := s.Files[sha256]
	if !exists {
		return fmt.Errorf("file with SHA256 %s not found", sha256)
	}
	if increment {
		meta.ModerationVotes++
	} else {
		meta.ModerationVotes--
	}
	meta.LastUpdated = time.Now().UTC() // TODO should this be with .UTC() also
	return nil                          // Commands handle saving
}

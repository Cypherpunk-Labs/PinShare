package p2p

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"pinshare/internal/store"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"

	// "github.com/libp2p/go-libp2p/core/discovery" // No longer directly used as an alias like discovery.Discovery
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	discovery_routing "github.com/libp2p/go-libp2p/p2p/discovery/routing"
)

type PubSubConfig struct {
	TopicAdvertiseInterval  time.Duration // Interval for finding peers
	AutoTopicDiscovery      bool
	EnablePeriodicPublish   bool
	PeriodicPublishInterval time.Duration
	TopicID                 string
}

type PubSubManager struct {
	ctx              context.Context
	ps               *pubsub.PubSub
	topic            *pubsub.Topic
	subscription     *pubsub.Subscription
	host             host.Host
	metadataStore    *store.MetadataStore
	dataFile         string
	routingDiscovery *discovery_routing.RoutingDiscovery
	psc              *PubSubConfig
}

func NewPubSubManager(ctx context.Context, h host.Host, kadDHT *dht.IpfsDHT, storeInstance *store.MetadataStore, dataFilePath string, config PubSubConfig) (*PubSubManager, error) {
	ps, err := pubsub.NewGossipSub(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub service: %w", err)
	}

	topic, err := ps.Join(config.TopicID)
	if err != nil {
		return nil, fmt.Errorf("failed to join topic %s: %w", config.TopicID, err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		topic.Close() // Clean up topic if subscription fails
		return nil, fmt.Errorf("failed to subscribe to topic %s: %w", config.TopicID, err)
	}

	manager := &PubSubManager{
		ctx:              ctx,
		ps:               ps,
		topic:            topic,
		subscription:     sub,
		host:             h,
		metadataStore:    storeInstance,
		dataFile:         dataFilePath,
		routingDiscovery: discovery_routing.NewRoutingDiscovery(kadDHT),
		psc:              &config,
	}

	go manager.handleIncomingMessages()

	if config.AutoTopicDiscovery {
		// Use a sensible default if TopicAdvertiseInterval is not set or too short
		discoveryInterval := config.TopicAdvertiseInterval
		if discoveryInterval <= 0 {
			discoveryInterval = 1 * time.Minute // Default peer discovery interval
		}
		go manager.DiscoverPeers(manager.ctx, discoveryInterval)
		fmt.Printf("[INFO] Automatic topic discovery initiated for %s with interval %v.\n", config.TopicID, discoveryInterval)
	}

	if config.EnablePeriodicPublish && config.PeriodicPublishInterval > 0 {
		go manager.periodicPublisher(ctx, config.PeriodicPublishInterval)
		fmt.Printf("[INFO] Periodic metadata publishing enabled with interval %v.\n", config.PeriodicPublishInterval)
	}

	return manager, nil
}

func (d *PubSubManager) DiscoverPeers(ctx context.Context, findPeersInterval time.Duration) {
	fmt.Printf("[PUBSUB] Starting discovery process for topic: %s\n", d.psc.TopicID)

	// // Advertise our presence on the topic via the DHT
	// fmt.Printf("[PUBSUB] Advertising presence for topic %s on DHT\n", MetadataTopicID)
	// advertiseTTL, err := d.routingDiscovery.Advertise(ctx, MetadataTopicID)
	// if err != nil {
	// 	fmt.Printf("[PUBSUB ERROR] Failed to advertise topic %s on DHT: %v\n", MetadataTopicID, err)
	// 	// Consider a retry mechanism or periodic re-advertisement based on TTL
	// } else {
	// 	fmt.Printf("[PUBSUB] Successfully advertised topic %s on DHT. TTL: %s. Will re-advertise if/when needed by libp2p or manually.\n", MetadataTopicID, advertiseTTL)
	// 	// For robust discovery, you might want to re-advertise periodically, e.g., every ttl/2.
	// }
	// fmt.Printf("[PUBSUB] Finished initial advertising attempt for topic %s\n", MetadataTopicID)

	const maxAdvertiseRetries = 5
	const advertiseRetryDelay = 10 * time.Second
	var advertiseErr error

	for i := 0; i < maxAdvertiseRetries; i++ {
		if ctx.Err() != nil {
			fmt.Printf("[PUBSUB] Context cancelled, aborting advertise for topic %s\n", d.psc.TopicID)
			return
		}

		fmt.Printf("[PUBSUB] Attempting to advertise topic %s (attempt %d/%d)\n", d.psc.TopicID, i+1, maxAdvertiseRetries)
		advertiseTTL, err := d.routingDiscovery.Advertise(ctx, d.psc.TopicID)
		if err == nil {
			fmt.Printf("[PUBSUB] Successfully advertised topic %s on DHT. TTL: %s.\n", d.psc.TopicID, advertiseTTL)
			advertiseErr = nil // Clear any previous error
			break              // Success
		}
		advertiseErr = err // Store the last error

		// Check for errors indicating no peers in the routing table.
		// Note: The exact error string might vary, or you might check for a specific error type like dht.ErrNoPeersInRoutingTable.
		if strings.Contains(err.Error(), "failed to find any peer in table") ||
			strings.Contains(err.Error(), "no peers to announce to") ||
			strings.Contains(err.Error(), "no route to peer") { // Adding another common variant
			fmt.Printf("[PUBSUB WARN] Failed to advertise topic %s (attempt %d/%d): %v. Retrying in %v...\n", d.psc.TopicID, i+1, maxAdvertiseRetries, err, advertiseRetryDelay)
			if i == maxAdvertiseRetries-1 {
				fmt.Printf("[PUBSUB ERROR] Failed to advertise topic %s after %d attempts: %v\n", d.psc.TopicID, maxAdvertiseRetries, err)
			}
			select {
			case <-time.After(advertiseRetryDelay):
				// Continue to next retry
			case <-ctx.Done():
				fmt.Printf("[PUBSUB] Context cancelled during advertise retry delay for topic %s\n", d.psc.TopicID)
				return
			}
		} else {
			fmt.Printf("[PUBSUB ERROR] Failed to advertise topic %s due to an unexpected error: %v\n", d.psc.TopicID, err)
			break // Don't retry for other types of errors
		}
	}
	if advertiseErr != nil {
		fmt.Printf("[PUBSUB] Finished advertising attempts for topic %s with error: %v\n", d.psc.TopicID, advertiseErr)
	} else {
		fmt.Printf("[PUBSUB] Finished initial advertising process for topic %s\n", d.psc.TopicID)
	}

	// Loop to continuously find new peers for the topic
	ticker := time.NewTicker(findPeersInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[PUBSUB] Stopping peer discovery for topic %s\n", d.psc.TopicID)
			return
		case <-ticker.C:
			fmt.Printf("[PUBSUB] Finding peers for topic: %s\n", d.psc.TopicID)
			peerChan, err := d.routingDiscovery.FindPeers(ctx, d.psc.TopicID)
			if err != nil {
				fmt.Printf("[PUBSUB ERROR] Failed to find peers for topic %s: %v\n", d.psc.TopicID, err)
				continue
			}

			for peerInfo := range peerChan {
				if peerInfo.ID == d.host.ID() || len(peerInfo.Addrs) == 0 {
					continue
				}

				if d.host.Network().Connectedness(peerInfo.ID) != network.Connected {
					fmt.Printf("[PUBSUB] Found new peer %s for topic %s. Attempting to connect.\n     multi-addr: %s \n", peerInfo.ID.String(), d.psc.TopicID, peerInfo.Addrs)
					if err := d.host.Connect(ctx, peerInfo); err != nil {
						// fmt.Printf("[PUBSUB WARN] Failed to connect to discovered peer %s: %v\n", peerInfo.ID.String(), err)
					} else {
						fmt.Printf("[PUBSUB] Successfully connected to peer %s for topic %s.\n", peerInfo.ID.String(), d.psc.TopicID)
					}
				}
			}
			fmt.Printf("[PUBSUB] Finished a round of finding peers for topic %s.\n", d.psc.TopicID)
		}
	}
}

func (psm *PubSubManager) handleIncomingMessages() {
	fmt.Printf("[INFO] Listening for messages on PubSub topic: %s\n", psm.psc.TopicID)
	for {
		select {
		case <-psm.ctx.Done():
			fmt.Println("[INFO] PubSub message handler shutting down.")
			if psm.subscription != nil {
				psm.subscription.Cancel()
			}
			if psm.topic != nil {
				psm.topic.Close()
			}
			return
		default:
			if psm.subscription == nil { // Subscription might have been cancelled
				time.Sleep(1 * time.Second) // Prevent tight loop if subscription is nil
				continue
			}
			msg, err := psm.subscription.Next(psm.ctx)
			if err != nil {
				if psm.ctx.Err() != nil {
					return
				}
				fmt.Printf("[ERROR] Failed to get next pubsub message: %v\n", err)
				// Consider specific error handling, e.g., context deadline exceeded vs other errors
				// If the error is persistent, could indicate a problem needing reconnection/re-subscription.
				// For now, continue and retry.
				if err.Error() == "context canceled" || err.Error() == "subscription cancelled" {
					fmt.Println("[INFO] PubSub subscription ended.")
					return
				}
				time.Sleep(1 * time.Second) // Avoid spamming errors
				continue
			}

			if msg.ReceivedFrom == psm.host.ID() {
				continue
			}

			fmt.Printf("[PUBSUB] Received message from %s (topic: %s)\n", msg.ReceivedFrom.String(), psm.psc.TopicID)

			// TODO: here we could begin to handle different message types over the same topic.
			var receivedMeta store.BaseMetadata
			if err := json.Unmarshal(msg.Data, &receivedMeta); err != nil {
				fmt.Printf("[ERROR] Failed to unmarshal received metadata from %s: %v\n", msg.ReceivedFrom.String(), err)
				continue
			}

			updated, err := psm.metadataStore.ApplyGossipUpdate(receivedMeta)
			if err != nil {
				fmt.Printf("[ERROR] Failed to apply gossiped update from %s for %s: %v\n", msg.ReceivedFrom.String(), receivedMeta.FileSHA256, err)
				continue
			}

			if updated {
				fmt.Printf("[INFO] Applied and saved update from peer %s for %s.\n", msg.ReceivedFrom.String(), receivedMeta.FileSHA256)
				if err := psm.metadataStore.Save(psm.dataFile); err != nil {
					fmt.Printf("[ERROR] Failed to save metadata after applying gossip update from %s: %v\n", msg.ReceivedFrom.String(), err)
				}
				newFile, err := ProcessDownload(receivedMeta)
				if err != nil {
					fmt.Printf("[ERROR] Failed to process download for %s: %v\n", receivedMeta.FileSHA256, err)
					continue
				}
				if newFile {
					fmt.Println("[INFO] Successfully pinned and cached download for " + receivedMeta.IPFSCID)
				} else {
					fmt.Println("[INFO] File did not pass security check for " + receivedMeta.IPFSCID)
				}
			}

		}
	}
}

func (psm *PubSubManager) PublishMetadata(metadata store.BaseMetadata) error {
	if psm.topic == nil {
		return fmt.Errorf("cannot publish: not joined to any topic")
	}

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata for publishing: %w", err)
	}

	fmt.Printf("[INFO] Publishing metadata for %s (LastUpdated: %s) to topic %s\n", metadata.FileSHA256, metadata.LastUpdated, psm.psc.TopicID)
	if psm.ctx.Err() != nil {
		return fmt.Errorf("cannot publish: context cancelled: %w", psm.ctx.Err())
	}
	return psm.topic.Publish(psm.ctx, jsonData)
}

func (psm *PubSubManager) ListPeers() []peer.ID {
	if psm.topic == nil || psm.ps == nil {
		return nil
	}
	return psm.ps.ListPeers(psm.psc.TopicID)
}

func (psm *PubSubManager) periodicPublisher(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Short initial delay to allow the node to connect and stabilize a bit
	select {
	case <-time.After(30 * time.Second): // Example initial delay
	case <-ctx.Done():
		return
	}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("[INFO] Stopping periodic metadata publisher.")
			return
		case <-ticker.C:
			if psm.metadataStore == nil {
				fmt.Fprintf(os.Stderr, "[WARN] Metadata store not available for periodic publishing.\n")
				continue
			}

			// allMetadata := psm.metadataStore.GetAllMetadata() // Using the new method from store
			allMetadata := psm.metadataStore.GetAllFiles()
			if len(allMetadata) == 0 {
				// fmt.Println("[INFO] No metadata to publish periodically.")
				continue
			}

			fmt.Printf("[INFO] Periodically publishing %d metadata entries...\n", len(allMetadata))
			for _, meta := range allMetadata {
				// if err := psm.PublishMetadata(meta); err != nil {
				if err := psm.PublishMetadata(*meta); err != nil { // [pubsub.go]
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to periodically publish metadata for %s: %v\n", meta.FileSHA256, err)
				}
				// Optional: Add a small delay between each publish if you have many items,
				// to avoid flooding the network or hitting rate limits.
				select {
				case <-time.After(100 * time.Millisecond): // Brief pause
				case <-ctx.Done(): // Check context in tight loops
					fmt.Println("[INFO] Periodic publishing interrupted.")
					return
				}
			}
			fmt.Printf("[INFO] Finished periodic publishing cycle for %d entries.\n", len(allMetadata))
		}
	}
}

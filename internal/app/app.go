package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pinshare/internal/api"
	"pinshare/internal/cmd"
	"pinshare/internal/config"
	"pinshare/internal/p2p"
	"pinshare/internal/psfs"
	"pinshare/internal/store"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

var (
	Node       host.Host
	P2PManager *p2p.PubSubManager
	kadDHT     *dht.IpfsDHT
)

func setupID(idFile string) crypto.PrivKey {
	var privKey crypto.PrivKey
	keyBytes, err := os.ReadFile(idFile)
	if err == nil {
		privKey, err = crypto.UnmarshalPrivateKey(keyBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error unmarshalling private key from %s: %v\n", idFile, err)
			os.Exit(1)
		}
		fmt.Printf("[INFO] Loaded identity from %s\n", idFile)
	} else if os.IsNotExist(err) {
		fmt.Printf("[INFO] Identity file %s not found, generating new identity...\n", idFile)
		privKey, _, err = crypto.GenerateKeyPair(crypto.Ed25519, -1)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error generating private key: %v\n", err)
			os.Exit(1)
		}
		keyBytes, err = crypto.MarshalPrivateKey(privKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error marshalling private key: %v\n", err)
			os.Exit(1)
		}
		if err = os.WriteFile(idFile, keyBytes, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Error saving private key to %s: %v\n", idFile, err)
			os.Exit(1)
		}
		fmt.Printf("[INFO] Saved new identity to %s\n", idFile)
	} else {
		fmt.Fprintf(os.Stderr, "[ERROR] Error reading identity file %s: %v\n", idFile, err)
		os.Exit(1)
	}
	return privKey
}

func Start() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var LoadedMetaData bool = false

	appconf, _ := config.LoadConfig()
	p2p.SetAppConfig(appconf)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n[INFO] Received shutdown signal, closing libp2p host and saving data...")
		if LoadedMetaData {
			if errStoreSave := store.GlobalStore.Save(appconf.MetaDataFile); errStoreSave != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Error saving data on exit: %v\n", errStoreSave)
			}
		}
		cancel()
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()

	// cmd.SetP2PManager(P2PManager)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] CLI Error: %s\n", err)
		if LoadedMetaData {
			if errSave := store.GlobalStore.Save(appconf.MetaDataFile); errSave != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Error saving data after command error: %v\n", errSave)
			}
		}
		os.Exit(1)
	}

	if len(os.Args) == 1 {
		fmt.Println("[INFO] No command given. Libp2p host is running. Press Ctrl+C to exit.")
		// ---------
		// start libp2p service here
		createFolders(appconf)

		err := store.GlobalStore.Load(appconf.MetaDataFile)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Warning: could not load data file '%s': %v\n", appconf.MetaDataFile, err)
			}
		}
		LoadedMetaData = true

		privKey := setupID(appconf.IdentityKeyFile)

		Node, err = p2p.NewHost(ctx, appconf.Libp2pPort, privKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to create libp2p host: %v\n", err)
			os.Exit(1)
		}
		defer Node.Close()

		// Set up the handler for direct messages
		p2p.SetDirectMessageHandler(Node)

		fmt.Printf("[INFO] Libp2p Host ID: %s\n", Node.ID())
		fmt.Println("[INFO] Libp2p Host Addresses:")
		for _, addr := range Node.Addrs() {
			fullAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", Node.ID().String()))
			peerAddr := addr.Encapsulate(fullAddr)
			fmt.Printf("  %s\n", peerAddr)
		}

		kadDHT, err = dht.New(ctx, Node)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to create DHT: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("[INFO] Bootstrapping DHT...")
		if err = kadDHT.Bootstrap(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "[WARNING] DHT bootstrap failed: %v. Node will attempt to discover peers through other means or retry.\n", err)
		} else {
			fmt.Println("[INFO] DHT bootstrap process initiated.")
		}

		pubSubConfig := p2p.PubSubConfig{
			TopicAdvertiseInterval:  30 * time.Second, // How often to run FindPeers loop
			AutoTopicDiscovery:      true,
			EnablePeriodicPublish:   true,
			PeriodicPublishInterval: 1 * time.Minute, // TODO decide on good period, set low for testing
			TopicID:                 appconf.MetadataTopicID,
		}

		P2PManager, err = p2p.NewPubSubManager(ctx, Node, kadDHT, store.GlobalStore, appconf.MetaDataFile, pubSubConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Failed to create PubSub Manager: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("[INFO] PubSub Manager initialized.")

		go func() {
			time.Sleep(1 * time.Second)
			fmt.Println("[INFO] Bootstrapping libp2p host against known peers (if any)...")
			p2p.Bootstrap(ctx, Node)
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if Node != nil && Node.Network() != nil && P2PManager != nil {
						fmt.Printf("\n[INFO] ----- Periodic Status Update -----\n")
						fmt.Printf("[INFO] Connected to %d peers.\n", len(Node.Network().Peers()))

						fmt.Println("[INFO] Host Addresses:")
						for _, addr := range Node.Addrs() {
							fullAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", Node.ID()))
							peerAddr := addr.Encapsulate(fullAddr)
							fmt.Printf("  %s\n", peerAddr)
						}

						topicPeers := P2PManager.ListPeers()
						if topicPeers != nil {
							fmt.Printf("[INFO] PubSub peers on metadata topic: %d (%v)\n", len(topicPeers), topicPeers)
						} else {
							fmt.Printf("[INFO] PubSub peers on metadata topic: 0 (manager or topic not fully initialized for listing)\n")
						}
						fmt.Printf("[INFO] ------------------------------------\n")
					}
				}
			}
		}()

		go startFileWatcher(ctx, appconf.UploadFolder, appconf.WatchInterval)

		if !appconf.FFArchiveNode {
			go bansetRoutine(ctx, appconf, appconf.WatchInterval) // TODO: ?define own config?
		}

		go func() {
			api.Start(ctx, Node)
			for {
				select {
				case <-ctx.Done():
					fmt.Println("[INFO] Stopping API.")
					return
				}
			}
		}()
		// end libp2p service here
		// ---------
		<-ctx.Done()
		fmt.Println("[INFO] Libp2p host shutting down.")
	}
}

func bansetRoutine(ctx context.Context, appconf *config.AppConfig, interval time.Duration) {
	fmt.Println("[INFO] Starting banset routine")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("[INFO] Running banset routine")
			Files := store.GlobalStore.GetAllFiles()
			for _, file := range Files {
				if file.BanSet > 0 {
					psfs.UnpinFileIPFS(file.IPFSCID)
				}
			}
		case <-ctx.Done():
			fmt.Println("[INFO] Stopping banset routine")
			return
		}
	}
}

func startFileWatcher(ctx context.Context, folderPath string, interval time.Duration) {
	fmt.Printf("[INFO] Starting file watcher for folder '%s' with interval %v\n", folderPath, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Printf("[INFO] Performing initial scan of '%s'...\n", folderPath)
	p2p.ProcessUploads(folderPath)

	for {
		select {
		case <-ticker.C:
			fmt.Printf("[INFO] Scanning '%s' for new files...\n", folderPath)
			p2p.ProcessUploads(folderPath)
		case <-ctx.Done():
			fmt.Println("[INFO] Stopping file watcher.")
			return
		}
	}
}

func createFolders(appconf *config.AppConfig) {
	if err := os.MkdirAll(appconf.UploadFolder, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Error creating upload directory '%s': %v\n", appconf.UploadFolder, err)
	}
	if err := os.MkdirAll(appconf.CacheFolder, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Error creating upload directory '%s': %v\n", appconf.CacheFolder, err)
	}
	if err := os.MkdirAll(appconf.RejectFolder, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Error creating reject directory '%s': %v\n", appconf.RejectFolder, err)
	}
}

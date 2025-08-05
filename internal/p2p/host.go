package p2p

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	// "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	// "[github.com/libp2p/go-libp2p/p2p/discovery/mdns](https://github.com/libp2p/go-libp2p/p2p/discovery/mdns)" // Optional: for local discovery
)

const DirectMessageProtocolID = "/pinshare/dm/1.0.0"

// NewHost creates a new libp2p host with DHT and attempts to bootstrap.
func NewHost(ctx context.Context, port int, privKey crypto.PrivKey) (host.Host, error) {
	// Check if port 50001 is in use. If so, increment until an open port is found.
	var dynport int = port
	fmt.Printf("[P2P-INFO] Testing if Port %d is in use. \n", port)
	for {
		addr := fmt.Sprintf("0.0.0.0:%d", dynport)
		conn, err := net.Listen("tcp", addr) // TODO: does not seem to conflict even if in use. no real value here.
		if err != nil {
			fmt.Printf("[P2P-INFO] Port %d is in use, trying next...\n", port)
			dynport++
			continue
		}
		conn.Close()
		break
	}

	listenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", dynport)
	// For QUIC (UDP), you might use:
	listenAddrUDP := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", dynport)

	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(listenAddr),    // Listen on TCP
		libp2p.ListenAddrStrings(listenAddrUDP), // Optionally listen on QUIC
		libp2p.DefaultSecurity,                  // Use default security transports (TLS, Noise)
		libp2p.DefaultMuxers,                    // Use default stream multiplexers (mplex, yamux)
		libp2p.NATPortMap(),                     // Attempt to open ports using uPNP for NATed environments
		libp2p.EnableHolePunching(),             // Enable hole punching for NAT traversal
		libp2p.EnableRelayService(),             // Enable circuit relay v2 service
		libp2p.EnableAutoNATv2(),                // Enable automatic NAT traversal
		libp2p.EnableRelay(),                    // Enable circuit relay v1 service
		// libp2p.EnableAutoRelay(),                // Use relays if the node is behind a NAT //BUG: deprecated
		// libp2p.EnableAutoRelayWithPeerSource() // TODO:
		// libp2p.EnableAutoRelayWithStaticRelays(), // TODO:

		// libp2p.EnableAutoRelayWithPeerSource(func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
		// 	peerChan := make(chan peer.AddrInfo)
		// 	go func() {
		// 		defer close(peerChan)
		// 		if kadDHT == nil {
		// 			//bug: only ever hits this section, need to pass in the DHT properly //TODO:
		// 			fmt.Println("[WARN] AutoRelay peer source called but DHT is not ready yet.")
		// 			return
		// 		}

		// 		routingDiscovery := discovery_routing.NewRoutingDiscovery(kadDHT)

		// 		relayTopic := MetadataTopicID                                                            // "/libp2p/circuit/relay/0.2.0/hop"                                          //
		// 		peerInfoCh, err := util.FindPeers(ctx, routingDiscovery, relayTopic, discovery.Limit(1)) //TODO: need different or no topic or use ipfs network to find relay
		// 		if err != nil {
		// 			fmt.Printf("[WARN] Failed to find peers for autorelay: %v\n", err)
		// 			return
		// 		}

		// 		for p := range peerInfoCh {
		// 			fmt.Printf("[DEBUG] Found %s peers for autorelay.\n", peerInfoCh[p].Addrs)
		// 			select {
		// 			case peerChan <- peerInfoCh[p]:
		// 			case <-ctx.Done():
		// 				return
		// 			}
		// 		}
		// 	}()
		// 	fmt.Println("[INFO] AutoRelay completed call.")
		// 	return peerChan
		// }),

		libp2p.EnableNATService(), // Help other peers discover their public address
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			kadDHT, err := dht.New(ctx, h, dht.Mode(dht.ModeAutoServer)) // was dht.ModeServer
			if err != nil {
				return nil, fmt.Errorf("failed to create DHT: %w", err)
			}
			return kadDHT, nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// fmt.Printf("[INFO] Libp2p host created with ID: %s\n", h.ID().String())
	// fmt.Println("[INFO] Host listening on addresses:")
	// for _, addr := range h.Addrs() {
	// 	fmt.Printf("  %s/p2p/%s\n", addr, h.ID().String())
	// }
	return h, nil
}

// Bootstrap connects to a set of bootstrap peers, primarily the IPFS default ones.
func Bootstrap(ctx context.Context, h host.Host) {
	// // kadDHT, ok := h.Routing().(*dht.IpfsDHT)
	// kadDHT, ok := h.Network().(*dht.IpfsDHT)
	// if !ok {
	// 	fmt.Println("[ERROR] Host routing is not Kademlia DHT, cannot bootstrap DHT.")
	// 	connectToDefaultBootstrapPeers(ctx, h)
	// 	return
	// }

	// fmt.Println("[INFO] Bootstrapping the DHT...")
	// if err := kadDHT.Bootstrap(ctx); err != nil {
	// 	fmt.Printf("[ERROR] DHT bootstrap failed: %v. Attempting manual connection to default peers.\n", err)
	// 	connectToDefaultBootstrapPeers(ctx, h)
	// 	return
	// }
	connectToDefaultBootstrapPeers(ctx, h)
	fmt.Println("[INFO] DHT bootstrap process initiated.")
	return
}

func connectToDefaultBootstrapPeers(ctx context.Context, h host.Host) {
	fmt.Println("[INFO] Connecting to default libp2p bootstrap peers...")
	var wg sync.WaitGroup
	for _, peerAddrStr := range dht.DefaultBootstrapPeers {
		// peerMA, err := multiaddr.NewMultiaddr(peerAddrStr)
		// if err != nil {
		// 	fmt.Printf("[WARN] Could not parse bootstrap peer multiaddr %s: %v\n", peerAddrStr, err)
		// 	continue
		// }
		// peerinfo, err := peer.AddrInfoFromP2pAddr(peerMA)
		peerinfo, err := peer.AddrInfoFromP2pAddr(peerAddrStr)
		if err != nil {
			fmt.Printf("[WARN] Could not get AddrInfo from bootstrap peer multiaddr %s: %v\n", peerAddrStr, err)
			continue
		}

		wg.Add(1)
		go func(pi peer.AddrInfo) {
			defer wg.Done()
			fmt.Printf("[INFO] Connecting to bootstrap peer: %s\n", pi.ID.String())
			if err := h.Connect(ctx, pi); err != nil {
				// fmt.Printf("[WARN] Failed to connect to bootstrap peer %s: %v\n", pi.ID.String(), err)
			} else {
				fmt.Printf("[INFO] Successfully connected to bootstrap peer: %s\n", pi.ID.String())
			}
		}(*peerinfo)
	}
	wg.Wait()
	fmt.Println("[INFO] Finished attempting to connect to bootstrap peers.")
}

func ParsePeerID(peerID string) (peer.ID, error) {
	return peer.Decode(peerID)
}

func DirectMessagePeer(ctx context.Context, h host.Host, peerID peer.ID, message []byte) error {
	s, err := h.NewStream(ctx, peerID, DirectMessageProtocolID)
	if err != nil {
		return fmt.Errorf("[DM ERROR] failed to open stream to peer %s: %w", peerID, err)
	}
	defer s.Close()

	_, err = s.Write(message)
	if err != nil {
		s.Reset() // Reset the stream on error
		return fmt.Errorf("[DM ERROR] failed to write to stream for peer %s: %w", peerID, err)
	}
	return nil
}

func SetDirectMessageHandler(h host.Host) {
	// The handler function for incoming streams.
	streamHandler := func(s network.Stream) {
		fmt.Printf("[DM] Received new direct message from %s\n", s.Conn().RemotePeer())
		defer s.Close()

		buf, err := io.ReadAll(s)
		if err != nil {
			fmt.Printf("[DM ERROR] Failed to read from direct message stream: %v\n", err)
			s.Reset()
			return
		}

		fmt.Printf("[DM] Message content: %s\n", string(buf)) //BUG // TODO: message is empty???

		_, err = s.Write([]byte("Message received!"))
		if err != nil {
			fmt.Printf("[DM ERROR] Failed to write response: %v\n", err) //TODO: BUG seems to hit here!
			s.Reset()
		}
	}
	h.SetStreamHandler(DirectMessageProtocolID, streamHandler)
	fmt.Printf("[INFO] Direct message handler registered for protocol: %s\n", DirectMessageProtocolID)
}

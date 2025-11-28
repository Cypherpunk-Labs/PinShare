package p2p

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"

	// "github.com/libp2p/go-libp2p-core/routing"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	libp2pquic "github.com/libp2p/go-libp2p/p2p/transport/quic"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	libp2pwebrtc "github.com/libp2p/go-libp2p/p2p/transport/webrtc"
	"github.com/libp2p/go-libp2p/p2p/transport/websocket"
	// ??p2p-circuit
	// "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	// "[github.com/libp2p/go-libp2p/p2p/discovery/mdns](https://github.com/libp2p/go-libp2p/p2p/discovery/mdns)" // Optional: for local discovery
)

const DirectMessageProtocolID = "/pinshare/dm/1.0.0"

// NewHost creates a new libp2p host with DHT and attempts to bootstrap.
func NewHost(ctx context.Context, port int, privKey crypto.PrivKey) (host.Host, error) {
	var kadDHT *dht.IpfsDHT
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

	ourlistenAddrs := []string{
		fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", dynport),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/webrtc-direct", dynport),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", dynport),
		fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", dynport),
		// "/ip4/0.0.0.0/tcp/4001",
		// "/ip6/::/tcp/4001",
		// "/ip4/0.0.0.0/udp/4001/webrtc-direct",
		// "/ip4/0.0.0.0/udp/4001/quic-v1",
		// "/ip4/0.0.0.0/udp/4001/quic-v1/webtransport",
		// "/ip6/::/udp/4001/webrtc-direct",
		// "/ip6/::/udp/4001/quic-v1",
		// "/ip6/::/udp/4001/quic-v1/webtransport",
	}

	// listenAddr := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", dynport)
	// // For QUIC (UDP), you might use:
	// listenAddrUDP := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", dynport)
	// listenAddrWebRTC := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/webrtc-direct", dynport)
	// listenAddrwebTransport := fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1/webtransport", dynport)

	// TODO: make a webcall to find our public IP then setup variables
	resp, err := http.Get("https://ifconfig.me/ip")
	if err != nil {
		fmt.Printf("[P2P-WARN] Could not get public IP: %v. May be behind a NAT.", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[P2P-WARN] Could not read public IP response: %v.", err)
	}
	publicIP := string(body)
	publicAddr := fmt.Sprintf("/ip4/%s/tcp/%d", publicIP, dynport)
	publicAddrUDP := fmt.Sprintf("/ip4/%s/udp/%d/quic-v1", publicIP, dynport)

	h, err := libp2p.New(
		// testing method to include public addr.
		// libp2p.AddrsFactory(func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		// 	return []multiaddr.Multiaddr{
		// 		multiaddr.StringCast(publicAddr),
		// 		multiaddr.StringCast(publicAddrUDP),
		// 	}
		// }), // using this explicitly only includes these addrs, the listenAddr ones are ignored.
		// // ok lets try adding these into the listenaddr instead

		// // opencode suggested
		// libp2p.AddrsFactory(func(addrs []multiaddr.Multiaddr) []multiaddr.Multiaddr {
		// 	var result []multiaddr.Multiaddr
		// 	for _, addr := range addrs {
		// 		// Filter out private addresses in container environments
		// 		if !isPrivateAddress(addr) {
		// 			result = append(result, addr)
		// 		}
		// 	}
		// 	// Add public announce addresses if configured
		// 	if len(announceAddrs) > 0 {
		// 		result = append(result, announceAddrs...)
		// 	}
		// 	return result
		// }),

		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(ourlistenAddrs...),
		// libp2p.ListenAddrStrings(listenAddrWebRTC),
		// libp2p.ListenAddrStrings(listenAddrwebTransport),
		// libp2p.ListenAddrStrings(listenAddr),    // Listen on TCP
		// libp2p.ListenAddrStrings(listenAddrUDP), // Optionally listen on QUIC
		libp2p.ListenAddrStrings(publicAddr),
		libp2p.ListenAddrStrings(publicAddrUDP),
		libp2p.DefaultSecurity,      // Use default security transports (TLS, Noise)
		libp2p.DefaultMuxers,        // Use default stream multiplexers (mplex, yamux)
		libp2p.NATPortMap(),         // Attempt to open ports using uPNP for NATed environments
		libp2p.EnableHolePunching(), // Enable hole punching for NAT traversal
		libp2p.EnableRelay(),        // Enable circuit relay v1 service
		libp2p.EnableRelayService(), // Enable circuit relay v2 service
		libp2p.EnableAutoNATv2(),    // Enable automatic NAT traversal
		// libp2p.EnableAutoRelay(),                // Use relays if the node is behind a NAT //BUG: deprecated
		// libp2p.EnableAutoRelayWithPeerSource(), // TODO:
		// libp2p.EnableAutoRelayWithStaticRelays(), // TODO:

		// // as per opencode
		// libp2p.EnableAutoRelayWithPeerSource(
		// 	func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
		// 		peerChan := make(chan peer.AddrInfo)
		// 		go func() {
		// 			defer close(peerChan)
		// 			// Use DHT to find relay peers
		// 			routingDiscovery := discovery_routing.NewRoutingDiscovery(kadDHT)
		// 			relayTopic := "/libp2p/circuit/relay/0.2.0/hop"
		// 			fmt.Printf("[INFO] AutoRelay1: Finding %d peers for rendezvous point %s\n", numPeers, relayTopic)
		// 			peerInfoCh, err := util.FindPeers(ctx, routingDiscovery, relayTopic, discovery.Limit(numPeers))
		// 			if err != nil {
		// 				return
		// 			}
		// 			foundPeers := 0
		// 			for _, p := range peerInfoCh {
		// 				if p.ID == "" {
		// 					continue
		// 				}
		// 				if foundPeers >= numPeers {
		// 					break
		// 				}
		// 				fmt.Printf("[DEBUG] AutoRelay1: Found potential relay peer: %s with addrs: %s\n", p.ID.String(), p.Addrs)
		// 				select {
		// 				case peerChan <- p:
		// 					foundPeers++
		// 				case <-ctx.Done():
		// 					return
		// 				}
		// 			}
		// 			fmt.Printf("[INFO] AutoRelay1: finished finding peers. Found %d.\n", foundPeers)
		// 		}()
		// 		return peerChan
		// 	},
		// 	autorelay.WithMinInterval(0),
		// ),

		// as per bryans suggestion
		libp2p.EnableAutoRelayWithPeerSource(
			func(ctx context.Context, numPeers int) <-chan peer.AddrInfo {
				return findRelayPeers(ctx, kadDHT, numPeers)
			},
			autorelay.WithMinCandidates(4),
			autorelay.WithMaxCandidates(8),
			autorelay.WithBootDelay(30*time.Second),
			autorelay.WithMinInterval(time.Minute),
		),

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
			var err error
			kadDHT, err = dht.New(ctx, h, dht.Mode(dht.ModeAutoServer)) // was dht.ModeServer
			if err != nil {
				return nil, fmt.Errorf("failed to create DHT: %w", err)
			}
			return kadDHT, nil
		}),
		libp2p.Transport(websocket.New),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Transport(libp2pquic.NewTransport),
		// libp2p.Transport(libp2prelay.New),
		libp2p.Transport(libp2pwebrtc.New),
		// libp2p.ShareTCPListener(), //deprecated
		// libp2p.Defaults,
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

// // TODO: test this works, it does not :(
// func findRelayPeers(ctx context.Context, kadDHT **dht.IpfsDHT, numPeers int) <-chan peer.AddrInfo {
// 	peerChan := make(chan peer.AddrInfo, numPeers)
// 	go func() {
// 		defer close(peerChan)

// 		if *kadDHT == nil {
// 			fmt.Println("[WARN] findRelayPeers called but DHT is not ready yet.")
// 			return
// 		}

// 		routingDiscovery := discovery_routing.NewRoutingDiscovery(*kadDHT)

// 		// We're looking for relays, so we advertise ourselves as needing a relay
// 		// and look for peers who provide the relay service.
// 		// The rendezvous point for circuit relay v2 is "/libp2p/relay"
// 		rendezvousPoint := "/libp2p/relay"
// 		fmt.Printf("[INFO] AutoRelay: Finding %d peers for rendezvous point %s\n", numPeers, rendezvousPoint)

// 		// FindPeers will return a channel of peers who are advertising the rendezvous point.
// 		peerInfoCh, err := discovery_util.FindPeers(ctx, routingDiscovery, rendezvousPoint)
// 		if err != nil {
// 			fmt.Printf("[WARN] Failed to find peers for autorelay: %v\n", err)
// 			return
// 		}

// 		foundPeers := 0
// 		for _, p := range peerInfoCh {
// 			if p.ID == "" {
// 				continue
// 			}
// 			if foundPeers >= numPeers {
// 				break
// 			}

// 			fmt.Printf("[DEBUG] AutoRelay: Found potential relay peer: %s with addrs: %s\n", p.ID.String(), p.Addrs)
// 			select {
// 			case peerChan <- p:
// 				foundPeers++
// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 		fmt.Printf("[INFO] AutoRelay: finished finding peers. Found %d.\n", foundPeers)
// 	}()
// 	return peerChan
// }

// bryans version
func findRelayPeers(ctx context.Context, kadDHT *dht.IpfsDHT, numPeers int) <-chan peer.AddrInfo {
	peerChan := make(chan peer.AddrInfo, numPeers)

	go func() {
		defer close(peerChan)

		// Wait for DHT to be ready
		if kadDHT == nil {
			fmt.Println("[RELAY] DHT not initialized yet, waiting...")
			time.Sleep(5 * time.Second)
			if kadDHT == nil {
				fmt.Println("[RELAY] DHT still not ready, aborting relay peer discovery")
				return
			}
		}

		fmt.Printf("[RELAY] Looking for %d relay candidates from DHT...\n", numPeers)

		// Get connected peers from the DHT routing table
		peers := kadDHT.RoutingTable().ListPeers()
		found := 0

		for _, p := range peers {
			if found >= numPeers {
				break
			}

			// Get peer's addresses from peerstore
			addrs := kadDHT.Host().Peerstore().Addrs(p)
			if len(addrs) > 0 {
				peerInfo := peer.AddrInfo{
					ID:    p,
					Addrs: addrs,
				}

				select {
				case peerChan <- peerInfo:
					found++
					fmt.Printf("[RELAY] Found relay candidate: %s\n", p.String())
				case <-ctx.Done():
					return
				}
			}
		}

		fmt.Printf("[RELAY] Relay peer discovery complete: found %d candidates\n", found)
	}()

	return peerChan
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

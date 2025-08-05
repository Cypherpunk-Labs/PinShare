# Tasks Pinshare

[] Application:
    [x] implement config feature flags
    [] CID validate and sha256 extract from multihash
    [] CID determine before upload (Chunking issues[static 256kb is used])
    [] Change from SHA256 to CID driven model.
    [] Cleanup comments
    [] Cleanup test func and redundant code
    [x] On startup check for all dependencies (similar task for security)
    [x] integrate IPFS node
        [x] add file
        [x] get cid/file
        [x] pin cid
        [x] unpid cid
    [x] fileWatcher from uploads directory for bulk imports to IPFS
    [] CRDT Store:
        [x] basemetadata
        [x] tagmetadata
        [x] voting
        [] Auth/fair voting
        [x] banset
    [] Network:
        [x] use libp2p
        [x] Set static node id with save/load
        [x] use kademila DHT
        [x] use pubsub gossip
            [x] handle messages
        [] encrypt gossip messages
        [] add direct message support
    [] CMD-CLI:
        [x] test chromeDP
        [x] test VT-WS lookup
        [x] test VT-WS submit
        [x] test store Tag Add
        [x] test store Tag Del
        [x] test store VoteUp
        [x] test store VoteDown
        [x] test store Tag VoteUp
        [x] test store Tag VoteDown
        [x] test store Ban
        [] test dependancy check
    [] API:
        [x] Create OpenAPI Spec
        [x] Generate code from spec
        [] web3auth 
            [] for voting
            [] for monitoring
        [] implement functions 
            [x] ListAllFiles
            [x] GetFileBySHA256
            [x] AddOrUpdateFile
            [] AddTag()
            [] RemoveTag()
            [] VoteForRemoval()
            [x] ListP2PPeers
            [x] ConnectToPeer
            [x] SendDirectMessage
                [] Has bug
            [x] GetP2PStatus
    [] Security:
        [x] Control Allowed file types
        [x] Validate filetype 
        [x] VT Pub WS Lookup (Brittle)
        [x] VT Pub WS Submit (Brittle)
        [x] Local Engine (clamAV) see https://gist.github.com/DerFichtl/d041785294d42259fa2b6ee4831c9a55
        [x] On startup test for capability log results and enable code-path preference.
        [] VT API Lookup Path
        [] VT API Submit Path
        [] P2P Malware/Hashset Service
        [] Network Banning (Abuse)
        [] Add virus total API funcs, enable paths if API key is set.
        [] Add lastresort AV scanner engine. [brew install clamav] read notes 
    [] Monitoring: 
        [x] Log Basics
        [] Collect logs
        [] send logs over p2p
        [x] (28/07/2025) enabled prometheus libp2p metrics on API Port /metrics
        [] implement custom metrics
        [] node health (CPU/RAM/storage free)
        [] network health 
[] DevOps: 
    [x] CI/CD - docker image build
    [] docker: 
        [x] create dockerfile
        [x] integrate ipfs node
        [x] integrate chromedp and chrome
        [x] test docker build v0.1.1
        [x] fix workdir '/opt/pinshare/data'
        [x] add clamav
        [x] test docker build v0.1.2
[] Documentation:
    [x] Readme.md Overview and Usage Guide
[] Testing:
    [] API: test all functions
    [] unit tests   

# Releases

Stage 1 (Preview):
    [x] Bulk import folder
    [x] Only Process PDF + Text
    [x] Pubsub basemetadata
    [x] No Auth
    [x] Static Node ID
    [x] Pinning Logic (pinset + banset)
    [x] security steps
        [x] Allowed file types
        [x] Filetype validation
        [x] VT-WS (chromeDP) last resort
        [x] Local Scan Engine (clamAV)



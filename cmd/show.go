package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

// StoryConfig represents the relevant fields in story.toml
type StoryConfig struct {
	Version            string    `toml:"version"`
	Network            string    `toml:"network"`
	EngineEndpoint     string    `toml:"engine-endpoint"`
	EngineJWTFile      string    `toml:"engine-jwt-file"`
	SnapshotInterval   int       `toml:"snapshot-interval"`
	SnapshotKeepRecent int       `toml:"snapshot-keep-recent"`
	MinRetainBlocks    int       `toml:"min-retain-blocks"`
	Pruning            string    `toml:"pruning"`
	AppDBBackend       string    `toml:"app-db-backend"`
	EVMBuildDelay      string    `toml:"evm-build-delay"`
	EVMBuildOptimistic bool      `toml:"evm-build-optimistic"`
	APIEnable          bool      `toml:"api-enable"`
	APIAddress         string    `toml:"api-address"`
	EnabledUnsafeCORS  bool      `toml:"enabled-unsafe-cors"`
	Log                LogConfig `toml:"log"`
}

// LogConfig represents the [log] section in story.toml
type LogConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
}

// ConfigConfig represents the relevant fields in config.toml
type ConfigConfig struct {
	ProxyApp               string `toml:"proxy_app"`
	Moniker                string `toml:"moniker"`
	DBDir                  string `toml:"db_dir"`
	GenesisFile            string `toml:"genesis_file"`
	PrivValidatorKeyFile   string `toml:"priv_validator_key_file"`
	PrivValidatorStateFile string `toml:"priv_validator_state_file"`
	NodeKeyFile            string `toml:"node_key_file"`
	ABCI                   string `toml:"abci"`
	FilterPeers            bool   `toml:"filter_peers"`

	RPC             RPCConfig             `toml:"rpc"`
	P2P             P2PConfig             `toml:"p2p"`
	Mempool         MempoolConfig         `toml:"mempool"`
	Statesync       StatesyncConfig       `toml:"statesync"`
	Blocksync       BlocksyncConfig       `toml:"blocksync"`
	Consensus       ConsensusConfig       `toml:"consensus"`
	Storage         StorageConfig         `toml:"storage"`
	TxIndex         TxIndexConfig         `toml:"tx_index"`
	Instrumentation InstrumentationConfig `toml:"instrumentation"`
}

// RPCConfig represents the [rpc] section in config.toml
type RPCConfig struct {
	Laddr                                string `toml:"laddr"`
	GRPCLaddr                            string `toml:"grpc_laddr"`
	GRPCMaxOpenConnections               int    `toml:"grpc_max_open_connections"`
	Unsafe                               bool   `toml:"unsafe"`
	MaxOpenConnections                   int    `toml:"max_open_connections"`
	MaxSubscriptionClients               int    `toml:"max_subscription_clients"`
	MaxSubscriptionsPerClient            int    `toml:"max_subscriptions_per_client"`
	ExperimentalSubscriptionBufferSize   int    `toml:"experimental_subscription_buffer_size"`
	ExperimentalWebsocketWriteBufferSize int    `toml:"experimental_websocket_write_buffer_size"`
	ExperimentalCloseOnSlowClient        bool   `toml:"experimental_close_on_slow_client"`
	TimeoutBroadcastTxCommit             string `toml:"timeout_broadcast_tx_commit"`
	MaxRequestBatchSize                  int    `toml:"max_request_batch_size"`
	MaxBodyBytes                         int    `toml:"max_body_bytes"`
	MaxHeaderBytes                       int    `toml:"max_header_bytes"`
	PPROFLaddr                           string `toml:"pprof_laddr"`
}

// P2PConfig represents the [p2p] section in config.toml
type P2PConfig struct {
	Laddr                        string `toml:"laddr"`
	ExternalAddress              string `toml:"external_address"`
	Seeds                        string `toml:"seeds"`
	PersistentPeers              string `toml:"persistent_peers"`
	AddrBookFile                 string `toml:"addr_book_file"`
	AddrBookStrict               bool   `toml:"addr_book_strict"`
	MaxNumInboundPeers           int    `toml:"max_num_inbound_peers"`
	MaxNumOutboundPeers          int    `toml:"max_num_outbound_peers"`
	UnconditionalPeerIDs         string `toml:"unconditional_peer_ids"`
	PersistentPeersMaxDialPeriod string `toml:"persistent_peers_max_dial_period"`
	FlushThrottleTimeout         string `toml:"flush_throttle_timeout"`
	MaxPacketMsgPayloadSize      int    `toml:"max_packet_msg_payload_size"`
	SendRate                     int    `toml:"send_rate"`
	RecvRate                     int    `toml:"recv_rate"`
	PEX                          bool   `toml:"pex"`
	SeedMode                     bool   `toml:"seed_mode"`
	AllowDuplicateIP             bool   `toml:"allow_duplicate_ip"`
	HandshakeTimeout             string `toml:"handshake_timeout"`
	DialTimeout                  string `toml:"dial_timeout"`
}

// MempoolConfig represents the [mempool] section in config.toml
type MempoolConfig struct {
	Type                  string `toml:"type"`
	Recheck               bool   `toml:"recheck"`
	RecheckTimeout        string `toml:"recheck_timeout"`
	Broadcast             bool   `toml:"broadcast"`
	Size                  int    `toml:"size"`
	MaxTxsBytes           int    `toml:"max_txs_bytes"`
	CacheSize             int    `toml:"cache_size"`
	KeepInvalidTxsInCache bool   `toml:"keep-invalid-txs-in-cache"`
	MaxTxBytes            int    `toml:"max_tx_bytes"`
	MaxBatchBytes         int    `toml:"max_batch_bytes"`
}

// StatesyncConfig represents the [statesync] section in config.toml
type StatesyncConfig struct {
	Enable              bool   `toml:"enable"`
	RPCServers          string `toml:"rpc_servers"`
	TrustHeight         int    `toml:"trust_height"`
	TrustHash           string `toml:"trust_hash"`
	TrustPeriod         string `toml:"trust_period"`
	DiscoveryTime       string `toml:"discovery_time"`
	TempDir             string `toml:"temp_dir"`
	ChunkRequestTimeout string `toml:"chunk_request_timeout"`
	ChunkFetchers       string `toml:"chunk_fetchers"`
}

// BlocksyncConfig represents the [blocksync] section in config.toml
type BlocksyncConfig struct {
	Version string `toml:"version"`
}

// ConsensusConfig represents the [consensus] section in config.toml
type ConsensusConfig struct {
	WALFile                     string `toml:"wal_file"`
	TimeoutPropose              string `toml:"timeout_propose"`
	TimeoutProposeDelta         string `toml:"timeout_propose_delta"`
	TimeoutPrevote              string `toml:"timeout_prevote"`
	TimeoutPrevoteDelta         string `toml:"timeout_prevote_delta"`
	TimeoutPrecommit            string `toml:"timeout_precommit"`
	TimeoutPrecommitDelta       string `toml:"timeout_precommit_delta"`
	TimeoutCommit               string `toml:"timeout_commit"`
	DoubleSignCheckHeight       int    `toml:"double_sign_check_height"`
	SkipTimeoutCommit           bool   `toml:"skip_timeout_commit"`
	CreateEmptyBlocks           bool   `toml:"create_empty_blocks"`
	CreateEmptyBlocksInterval   string `toml:"create_empty_blocks_interval"`
	PeerGossipSleepDuration     string `toml:"peer_gossip_sleep_duration"`
	PeerQueryMaj23SleepDuration string `toml:"peer_query_maj23_sleep_duration"`
}

// StorageConfig represents the [storage] section in config.toml
type StorageConfig struct {
	DiscardABCIResponses bool `toml:"discard_abci_responses"`
}

// TxIndexConfig represents the [tx_index] section in config.toml
type TxIndexConfig struct {
	Indexer  string `toml:"indexer"`
	PSQLConn string `toml:"psql-conn"`
}

// InstrumentationConfig represents the [instrumentation] section in config.toml
type InstrumentationConfig struct {
	Prometheus           bool   `toml:"prometheus"`
	PrometheusListenAddr string `toml:"prometheus_listen_addr"`
	MaxOpenConnections   int    `toml:"max_open_connections"`
	Namespace            string `toml:"namespace"`
}

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display configuration values from Story and CometBFT services",
	Long:  `Provides subcommands to display specific configuration values from story.toml and config.toml.`,
}

// Subcommands for story.toml
var (
	showSnapshotIntervalCmd = &cobra.Command{
		Use:   "snapshot-interval",
		Short: "Show the snapshot-interval value from story.toml",
		RunE:  runSnapshotInterval,
	}

	showSnapshotKeepRecentCmd = &cobra.Command{
		Use:   "snapshot-keep-recent",
		Short: "Show the snapshot-keep-recent value from story.toml",
		RunE:  runSnapshotKeepRecent,
	}

	showMinRetainBlocksCmd = &cobra.Command{
		Use:   "min-retain-blocks",
		Short: "Show the min-retain-blocks value from story.toml",
		RunE:  runMinRetainBlocks,
	}

	showPruningCmd = &cobra.Command{
		Use:   "pruning-mode",
		Short: "Show the pruning mode value from story.toml",
		RunE:  runPruningMode,
	}

	showAppDBBackendCmd = &cobra.Command{
		Use:   "app-db-backend",
		Short: "Show the app-db-backend value from story.toml",
		RunE:  runAppDBBackend,
	}

	showEVMBuildDelayCmd = &cobra.Command{
		Use:   "evm-build-delay",
		Short: "Show the evm-build-delay value from story.toml",
		RunE:  runEVMBuildDelay,
	}

	showEVMBuildOptimisticCmd = &cobra.Command{
		Use:   "evm-build-optimistic",
		Short: "Show the evm-build-optimistic value from story.toml",
		RunE:  runEVMBuildOptimistic,
	}

	showAPIEnableCmd = &cobra.Command{
		Use:   "api-enable",
		Short: "Show the api-enable value from story.toml",
		RunE:  runAPIEnable,
	}

	showAPIAddressCmd = &cobra.Command{
		Use:   "api-address",
		Short: "Show the api-address value from story.toml",
		RunE:  runAPIAddress,
	}

	showEnabledUnsafeCORSCmd = &cobra.Command{
		Use:   "enabled-unsafe-cors",
		Short: "Show the enabled-unsafe-cors value from story.toml",
		RunE:  runEnabledUnsafeCORS,
	}

	showLogLevelCmd = &cobra.Command{
		Use:   "log-level",
		Short: "Show the log level from story.toml",
		RunE:  runLogLevel,
	}

	showLogFormatCmd = &cobra.Command{
		Use:   "log-format",
		Short: "Show the log format from story.toml",
		RunE:  runLogFormat,
	}
)

// Subcommands for config.toml
var (
	showProxyAppCmd = &cobra.Command{
		Use:   "proxy-app",
		Short: "Show the proxy_app value from config.toml",
		RunE:  runProxyApp,
	}

	showMonikerCmd = &cobra.Command{
		Use:   "moniker",
		Short: "Show the moniker value from config.toml",
		RunE:  runMoniker,
	}

	showDBDirCmd = &cobra.Command{
		Use:   "db-dir",
		Short: "Show the db_dir value from config.toml",
		RunE:  runDBDir,
	}

	showGenesisFileCmd = &cobra.Command{
		Use:   "genesis-file",
		Short: "Show the genesis_file value from config.toml",
		RunE:  runGenesisFile,
	}

	showPrivValidatorKeyFileCmd = &cobra.Command{
		Use:   "priv-validator-key-file",
		Short: "Show the priv_validator_key_file value from config.toml",
		RunE:  runPrivValidatorKeyFile,
	}

	showPrivValidatorStateFileCmd = &cobra.Command{
		Use:   "priv-validator-state-file",
		Short: "Show the priv_validator_state_file value from config.toml",
		RunE:  runPrivValidatorStateFile,
	}

	showNodeKeyFileCmd = &cobra.Command{
		Use:   "node-key-file",
		Short: "Show the node_key_file value from config.toml",
		RunE:  runNodeKeyFile,
	}

	showABCICommand = &cobra.Command{
		Use:   "abci",
		Short: "Show the abci value from config.toml",
		RunE:  runABCI,
	}

	showFilterPeersCmd = &cobra.Command{
		Use:   "filter-peers",
		Short: "Show the filter_peers value from config.toml",
		RunE:  runFilterPeers,
	}

	// [rpc] section
	showRPCLaddrCmd = &cobra.Command{
		Use:   "rpc-laddr",
		Short: "Show the rpc.laddr value from config.toml",
		RunE:  runRPCLaddr,
	}

	showRPCGRPCLaddrCmd = &cobra.Command{
		Use:   "rpc-grpc-laddr",
		Short: "Show the rpc.grpc_laddr value from config.toml",
		RunE:  runRPCGRPCLaddr,
	}

	showRPCMaxOpenConnectionsCmd = &cobra.Command{
		Use:   "rpc-grpc-max-open-connections",
		Short: "Show the rpc.grpc_max_open_connections value from config.toml",
		RunE:  runRPCMaxOpenConnections,
	}

	showRPCUnsafeCmd = &cobra.Command{
		Use:   "rpc-unsafe",
		Short: "Show the rpc.unsafe value from config.toml",
		RunE:  runRPCUnsafe,
	}

	showRPCMaxSubscriptionClientsCmd = &cobra.Command{
		Use:   "rpc-max-subscription-clients",
		Short: "Show the rpc.max_subscription_clients value from config.toml",
		RunE:  runRPCMaxSubscriptionClients,
	}

	showRPCMaxSubscriptionsPerClientCmd = &cobra.Command{
		Use:   "rpc-max-subscriptions-per-client",
		Short: "Show the rpc.max_subscriptions_per_client value from config.toml",
		RunE:  runRPCMaxSubscriptionsPerClient,
	}

	showRPCExperimentalSubscriptionBufferSizeCmd = &cobra.Command{
		Use:   "rpc-experimental-subscription-buffer-size",
		Short: "Show the rpc.experimental_subscription_buffer_size value from config.toml",
		RunE:  runRPCExperimentalSubscriptionBufferSize,
	}

	showRPCExperimentalWebsocketWriteBufferSizeCmd = &cobra.Command{
		Use:   "rpc-experimental-websocket-write-buffer-size",
		Short: "Show the rpc.experimental_websocket_write_buffer_size value from config.toml",
		RunE:  runRPCExperimentalWebsocketWriteBufferSize,
	}

	showRPCExperimentalCloseOnSlowClientCmd = &cobra.Command{
		Use:   "rpc-experimental-close-on-slow-client",
		Short: "Show the rpc.experimental_close_on_slow_client value from config.toml",
		RunE:  runRPCExperimentalCloseOnSlowClient,
	}

	showRPCTimeoutBroadcastTxCommitCmd = &cobra.Command{
		Use:   "rpc-timeout-broadcast-tx-commit",
		Short: "Show the rpc.timeout_broadcast_tx_commit value from config.toml",
		RunE:  runRPCTimeoutBroadcastTxCommit,
	}

	showRPCMaxRequestBatchSizeCmd = &cobra.Command{
		Use:   "rpc-max-request-batch-size",
		Short: "Show the rpc.max_request_batch_size value from config.toml",
		RunE:  runRPCMaxRequestBatchSize,
	}

	showRPCMaxBodyBytesCmd = &cobra.Command{
		Use:   "rpc-max-body-bytes",
		Short: "Show the rpc.max_body_bytes value from config.toml",
		RunE:  runRPCMaxBodyBytes,
	}

	showRPCMaxHeaderBytesCmd = &cobra.Command{
		Use:   "rpc-max-header-bytes",
		Short: "Show the rpc.max_header_bytes value from config.toml",
		RunE:  runRPCMaxHeaderBytes,
	}

	showRPCPPROFLaddrCmd = &cobra.Command{
		Use:   "rpc-pprof-laddr",
		Short: "Show the rpc.pprof_laddr value from config.toml",
		RunE:  runRPCPPROFLaddr,
	}

	// [p2p] section
	showP2PLaddrCmd = &cobra.Command{
		Use:   "p2p-laddr",
		Short: "Show the p2p.laddr value from config.toml",
		RunE:  runP2PLaddr,
	}

	showP2PExternalAddressCmd = &cobra.Command{
		Use:   "p2p-external-address",
		Short: "Show the p2p.external_address value from config.toml",
		RunE:  runP2PExternalAddress,
	}

	showP2PSeedsCmd = &cobra.Command{
		Use:   "p2p-seeds",
		Short: "Show the p2p.seeds value from config.toml",
		RunE:  runP2PSeeds,
	}

	showP2PPersistentPeersCmd = &cobra.Command{
		Use:   "p2p-persistent-peers",
		Short: "Show the p2p.persistent_peers value from config.toml",
		RunE:  runP2PPersistentPeers,
	}

	showP2PAddrBookFileCmd = &cobra.Command{
		Use:   "p2p-addr-book-file",
		Short: "Show the p2p.addr_book_file value from config.toml",
		RunE:  runP2PAddrBookFile,
	}

	showP2PAddrBookStrictCmd = &cobra.Command{
		Use:   "p2p-addr-book-strict",
		Short: "Show the p2p.addr_book_strict value from config.toml",
		RunE:  runP2PAddrBookStrict,
	}

	showP2PMaxNumInboundPeersCmd = &cobra.Command{
		Use:   "p2p-max-num-inbound-peers",
		Short: "Show the p2p.max_num_inbound_peers value from config.toml",
		RunE:  runP2PMaxNumInboundPeers,
	}

	showP2PMaxNumOutboundPeersCmd = &cobra.Command{
		Use:   "p2p-max-num-outbound-peers",
		Short: "Show the p2p.max_num_outbound_peers value from config.toml",
		RunE:  runP2PMaxNumOutboundPeers,
	}

	showP2PUnconditionalPeerIDsCmd = &cobra.Command{
		Use:   "p2p-unconditional-peer-ids",
		Short: "Show the p2p.unconditional_peer_ids value from config.toml",
		RunE:  runP2PUnconditionalPeerIDs,
	}

	showP2PPersistentPeersMaxDialPeriodCmd = &cobra.Command{
		Use:   "p2p-persistent-peers-max-dial-period",
		Short: "Show the p2p.persistent_peers_max_dial_period value from config.toml",
		RunE:  runP2PPersistentPeersMaxDialPeriod,
	}

	showP2PFlushThrottleTimeoutCmd = &cobra.Command{
		Use:   "p2p-flush-throttle-timeout",
		Short: "Show the p2p.flush_throttle_timeout value from config.toml",
		RunE:  runP2PFlushThrottleTimeout,
	}

	showP2PMaxPacketMsgPayloadSizeCmd = &cobra.Command{
		Use:   "p2p-max-packet-msg-payload-size",
		Short: "Show the p2p.max_packet_msg_payload_size value from config.toml",
		RunE:  runP2PMaxPacketMsgPayloadSize,
	}

	showP2PSendRateCmd = &cobra.Command{
		Use:   "p2p-send-rate",
		Short: "Show the p2p.send_rate value from config.toml",
		RunE:  runP2PSendRate,
	}

	showP2PRecvRateCmd = &cobra.Command{
		Use:   "p2p-recv-rate",
		Short: "Show the p2p.recv_rate value from config.toml",
		RunE:  runP2PRecvRate,
	}

	showP2PPEXCmd = &cobra.Command{
		Use:   "p2p-pex",
		Short: "Show the p2p.pex value from config.toml",
		RunE:  runP2PPEX,
	}

	showP2PSeedModeCmd = &cobra.Command{
		Use:   "p2p-seed-mode",
		Short: "Show the p2p.seed_mode value from config.toml",
		RunE:  runP2PSeedMode,
	}

	showP2PAllowDuplicateIPCmd = &cobra.Command{
		Use:   "p2p-allow-duplicate-ip",
		Short: "Show the p2p.allow_duplicate_ip value from config.toml",
		RunE:  runP2PAllowDuplicateIP,
	}

	showP2PHandshakeTimeoutCmd = &cobra.Command{
		Use:   "p2p-handshake-timeout",
		Short: "Show the p2p.handshake_timeout value from config.toml",
		RunE:  runP2PHandshakeTimeout,
	}

	showP2PDialTimeoutCmd = &cobra.Command{
		Use:   "p2p-dial-timeout",
		Short: "Show the p2p.dial_timeout value from config.toml",
		RunE:  runP2PDialTimeout,
	}

	// [mempool] section
	showMempoolTypeCmd = &cobra.Command{
		Use:   "mempool-type",
		Short: "Show the mempool.type value from config.toml",
		RunE:  runMempoolType,
	}

	showMempoolRecheckCmd = &cobra.Command{
		Use:   "mempool-recheck",
		Short: "Show the mempool.recheck value from config.toml",
		RunE:  runMempoolRecheck,
	}

	showMempoolRecheckTimeoutCmd = &cobra.Command{
		Use:   "mempool-recheck-timeout",
		Short: "Show the mempool.recheck_timeout value from config.toml",
		RunE:  runMempoolRecheckTimeout,
	}

	showMempoolBroadcastCmd = &cobra.Command{
		Use:   "mempool-broadcast",
		Short: "Show the mempool.broadcast value from config.toml",
		RunE:  runMempoolBroadcast,
	}

	showMempoolSizeCmd = &cobra.Command{
		Use:   "mempool-size",
		Short: "Show the mempool.size value from config.toml",
		RunE:  runMempoolSize,
	}

	showMempoolMaxTxsBytesCmd = &cobra.Command{
		Use:   "mempool-max-txs-bytes",
		Short: "Show the mempool.max_txs_bytes value from config.toml",
		RunE:  runMempoolMaxTxsBytes,
	}

	showMempoolCacheSizeCmd = &cobra.Command{
		Use:   "mempool-cache-size",
		Short: "Show the mempool.cache_size value from config.toml",
		RunE:  runMempoolCacheSize,
	}

	showMempoolKeepInvalidTxsInCacheCmd = &cobra.Command{
		Use:   "mempool-keep-invalid-txs-in-cache",
		Short: "Show the mempool.keep-invalid-txs-in-cache value from config.toml",
		RunE:  runMempoolKeepInvalidTxsInCache,
	}

	showMempoolMaxTxBytesCmd = &cobra.Command{
		Use:   "mempool-max-tx-bytes",
		Short: "Show the mempool.max_tx_bytes value from config.toml",
		RunE:  runMempoolMaxTxBytes,
	}

	showMempoolMaxBatchBytesCmd = &cobra.Command{
		Use:   "mempool-max-batch-bytes",
		Short: "Show the mempool.max_batch_bytes value from config.toml",
		RunE:  runMempoolMaxBatchBytes,
	}

	// [statesync] section
	showStatesyncEnableCmd = &cobra.Command{
		Use:   "statesync-enable",
		Short: "Show the statesync.enable value from config.toml",
		RunE:  runStatesyncEnable,
	}

	showStatesyncRPCServersCmd = &cobra.Command{
		Use:   "statesync-rpc-servers",
		Short: "Show the statesync.rpc_servers value from config.toml",
		RunE:  runStatesyncRPCServers,
	}

	showStatesyncTrustHeightCmd = &cobra.Command{
		Use:   "statesync-trust-height",
		Short: "Show the statesync.trust_height value from config.toml",
		RunE:  runStatesyncTrustHeight,
	}

	showStatesyncTrustHashCmd = &cobra.Command{
		Use:   "statesync-trust-hash",
		Short: "Show the statesync.trust_hash value from config.toml",
		RunE:  runStatesyncTrustHash,
	}

	showStatesyncTrustPeriodCmd = &cobra.Command{
		Use:   "statesync-trust-period",
		Short: "Show the statesync.trust_period value from config.toml",
		RunE:  runStatesyncTrustPeriod,
	}

	showStatesyncDiscoveryTimeCmd = &cobra.Command{
		Use:   "statesync-discovery-time",
		Short: "Show the statesync.discovery_time value from config.toml",
		RunE:  runStatesyncDiscoveryTime,
	}

	showStatesyncTempDirCmd = &cobra.Command{
		Use:   "statesync-temp-dir",
		Short: "Show the statesync.temp_dir value from config.toml",
		RunE:  runStatesyncTempDir,
	}

	showStatesyncChunkRequestTimeoutCmd = &cobra.Command{
		Use:   "statesync-chunk-request-timeout",
		Short: "Show the statesync.chunk_request_timeout value from config.toml",
		RunE:  runStatesyncChunkRequestTimeout,
	}

	showStatesyncChunkFetchersCmd = &cobra.Command{
		Use:   "statesync-chunk-fetchers",
		Short: "Show the statesync.chunk_fetchers value from config.toml",
		RunE:  runStatesyncChunkFetchers,
	}

	// [blocksync] section
	showBlocksyncVersionCmd = &cobra.Command{
		Use:   "blocksync-version",
		Short: "Show the blocksync.version value from config.toml",
		RunE:  runBlocksyncVersion,
	}

	// [consensus] section
	showConsensusWALFileCmd = &cobra.Command{
		Use:   "consensus-wal-file",
		Short: "Show the consensus.wal_file value from config.toml",
		RunE:  runConsensusWALFile,
	}

	showConsensusTimeoutProposeCmd = &cobra.Command{
		Use:   "consensus-timeout-propose",
		Short: "Show the consensus.timeout_propose value from config.toml",
		RunE:  runConsensusTimeoutPropose,
	}

	showConsensusTimeoutProposeDeltaCmd = &cobra.Command{
		Use:   "consensus-timeout-propose-delta",
		Short: "Show the consensus.timeout_propose_delta value from config.toml",
		RunE:  runConsensusTimeoutProposeDelta,
	}

	showConsensusTimeoutPrevoteCmd = &cobra.Command{
		Use:   "consensus-timeout-prevote",
		Short: "Show the consensus.timeout_prevote value from config.toml",
		RunE:  runConsensusTimeoutPrevote,
	}

	showConsensusTimeoutPrevoteDeltaCmd = &cobra.Command{
		Use:   "consensus-timeout-prevote-delta",
		Short: "Show the consensus.timeout_prevote_delta value from config.toml",
		RunE:  runConsensusTimeoutPrevoteDelta,
	}

	showConsensusTimeoutPrecommitCmd = &cobra.Command{
		Use:   "consensus-timeout-precommit",
		Short: "Show the consensus.timeout_precommit value from config.toml",
		RunE:  runConsensusTimeoutPrecommit,
	}

	showConsensusTimeoutPrecommitDeltaCmd = &cobra.Command{
		Use:   "consensus-timeout-precommit-delta",
		Short: "Show the consensus.timeout_precommit_delta value from config.toml",
		RunE:  runConsensusTimeoutPrecommitDelta,
	}

	showConsensusTimeoutCommitCmd = &cobra.Command{
		Use:   "consensus-timeout-commit",
		Short: "Show the consensus.timeout_commit value from config.toml",
		RunE:  runConsensusTimeoutCommit,
	}

	showConsensusDoubleSignCheckHeightCmd = &cobra.Command{
		Use:   "consensus-double-sign-check-height",
		Short: "Show the consensus.double_sign_check_height value from config.toml",
		RunE:  runConsensusDoubleSignCheckHeight,
	}

	showConsensusSkipTimeoutCommitCmd = &cobra.Command{
		Use:   "consensus-skip-timeout-commit",
		Short: "Show the consensus.skip_timeout_commit value from config.toml",
		RunE:  runConsensusSkipTimeoutCommit,
	}

	showConsensusCreateEmptyBlocksCmd = &cobra.Command{
		Use:   "consensus-create-empty-blocks",
		Short: "Show the consensus.create_empty_blocks value from config.toml",
		RunE:  runConsensusCreateEmptyBlocks,
	}

	showConsensusCreateEmptyBlocksIntervalCmd = &cobra.Command{
		Use:   "consensus-create-empty-blocks-interval",
		Short: "Show the consensus.create_empty_blocks_interval value from config.toml",
		RunE:  runConsensusCreateEmptyBlocksInterval,
	}

	showConsensusPeerGossipSleepDurationCmd = &cobra.Command{
		Use:   "consensus-peer-gossip-sleep-duration",
		Short: "Show the consensus.peer_gossip_sleep_duration value from config.toml",
		RunE:  runConsensusPeerGossipSleepDuration,
	}

	showConsensusPeerQueryMaj23SleepDurationCmd = &cobra.Command{
		Use:   "consensus-peer-query-maj23-sleep-duration",
		Short: "Show the consensus.peer_query_maj23_sleep_duration value from config.toml",
		RunE:  runConsensusPeerQueryMaj23SleepDuration,
	}

	// [storage] section
	showStorageDiscardABCIResponsesCmd = &cobra.Command{
		Use:   "storage-discard-abci-responses",
		Short: "Show the storage.discard_abci_responses value from config.toml",
		RunE:  runStorageDiscardABCIResponses,
	}

	// [tx_index] section
	showTxIndexIndexerCmd = &cobra.Command{
		Use:   "tx-index-indexer",
		Short: "Show the tx_index.indexer value from config.toml",
		RunE:  runTxIndexIndexer,
	}

	showTxIndexPSQLConnCmd = &cobra.Command{
		Use:   "tx-index-psql-conn",
		Short: "Show the tx_index.psql-conn value from config.toml",
		RunE:  runTxIndexPSQLConn,
	}

	// [instrumentation] section
	showInstrumentationPrometheusCmd = &cobra.Command{
		Use:   "instrumentation-prometheus",
		Short: "Show the instrumentation.prometheus value from config.toml",
		RunE:  runInstrumentationPrometheus,
	}

	showInstrumentationPrometheusListenAddrCmd = &cobra.Command{
		Use:   "instrumentation-prometheus-listen-addr",
		Short: "Show the instrumentation.prometheus_listen_addr value from config.toml",
		RunE:  runInstrumentationPrometheusListenAddr,
	}

	showInstrumentationMaxOpenConnectionsCmd = &cobra.Command{
		Use:   "instrumentation-max-open-connections",
		Short: "Show the instrumentation.max_open_connections value from config.toml",
		RunE:  runInstrumentationMaxOpenConnections,
	}

	showInstrumentationNamespaceCmd = &cobra.Command{
		Use:   "instrumentation-namespace",
		Short: "Show the instrumentation.namespace value from config.toml",
		RunE:  runInstrumentationNamespace,
	}
)

// init fonksiyonunu güncelleyin
func init() {
	// Add 'show' command to root
	rootCmd.AddCommand(showCmd)

	// Add story.toml subcommands
	showCmd.AddCommand(showSnapshotIntervalCmd)
	showCmd.AddCommand(showSnapshotKeepRecentCmd)
	showCmd.AddCommand(showMinRetainBlocksCmd)
	showCmd.AddCommand(showPruningCmd)
	showCmd.AddCommand(showAppDBBackendCmd)
	showCmd.AddCommand(showEVMBuildDelayCmd)
	showCmd.AddCommand(showEVMBuildOptimisticCmd)
	showCmd.AddCommand(showAPIEnableCmd)
	showCmd.AddCommand(showAPIAddressCmd)
	showCmd.AddCommand(showEnabledUnsafeCORSCmd)
	showCmd.AddCommand(showLogLevelCmd)
	showCmd.AddCommand(showLogFormatCmd)

	// Add config.toml subcommands
	showCmd.AddCommand(showProxyAppCmd)
	showCmd.AddCommand(showMonikerCmd)
	showCmd.AddCommand(showDBDirCmd)
	showCmd.AddCommand(showGenesisFileCmd)
	showCmd.AddCommand(showPrivValidatorKeyFileCmd)
	showCmd.AddCommand(showPrivValidatorStateFileCmd)
	showCmd.AddCommand(showNodeKeyFileCmd)
	showCmd.AddCommand(showABCICommand)
	showCmd.AddCommand(showFilterPeersCmd)

	// Add newly added config.toml subcommands
	showCmd.AddCommand(showRPCLaddrCmd)
	showCmd.AddCommand(showRPCGRPCLaddrCmd)
	showCmd.AddCommand(showRPCMaxOpenConnectionsCmd)
	showCmd.AddCommand(showRPCUnsafeCmd)
	showCmd.AddCommand(showRPCMaxSubscriptionClientsCmd)
	showCmd.AddCommand(showRPCMaxSubscriptionsPerClientCmd)
	showCmd.AddCommand(showRPCExperimentalSubscriptionBufferSizeCmd)
	showCmd.AddCommand(showRPCExperimentalWebsocketWriteBufferSizeCmd)
	showCmd.AddCommand(showRPCExperimentalCloseOnSlowClientCmd)
	showCmd.AddCommand(showRPCTimeoutBroadcastTxCommitCmd)
	showCmd.AddCommand(showRPCMaxRequestBatchSizeCmd)
	showCmd.AddCommand(showRPCMaxBodyBytesCmd)
	showCmd.AddCommand(showRPCMaxHeaderBytesCmd)
	showCmd.AddCommand(showRPCPPROFLaddrCmd)

	showCmd.AddCommand(showP2PLaddrCmd)
	showCmd.AddCommand(showP2PExternalAddressCmd)
	showCmd.AddCommand(showP2PSeedsCmd)
	showCmd.AddCommand(showP2PPersistentPeersCmd)
	showCmd.AddCommand(showP2PAddrBookFileCmd)
	showCmd.AddCommand(showP2PAddrBookStrictCmd)
	showCmd.AddCommand(showP2PMaxNumInboundPeersCmd)
	showCmd.AddCommand(showP2PMaxNumOutboundPeersCmd)
	showCmd.AddCommand(showP2PUnconditionalPeerIDsCmd)
	showCmd.AddCommand(showP2PPersistentPeersMaxDialPeriodCmd)
	showCmd.AddCommand(showP2PFlushThrottleTimeoutCmd)
	showCmd.AddCommand(showP2PMaxPacketMsgPayloadSizeCmd)
	showCmd.AddCommand(showP2PSendRateCmd)
	showCmd.AddCommand(showP2PRecvRateCmd)
	showCmd.AddCommand(showP2PPEXCmd)
	showCmd.AddCommand(showP2PSeedModeCmd)
	showCmd.AddCommand(showP2PAllowDuplicateIPCmd)
	showCmd.AddCommand(showP2PHandshakeTimeoutCmd)
	showCmd.AddCommand(showP2PDialTimeoutCmd)

	showCmd.AddCommand(showMempoolTypeCmd)
	showCmd.AddCommand(showMempoolRecheckCmd)
	showCmd.AddCommand(showMempoolRecheckTimeoutCmd)
	showCmd.AddCommand(showMempoolBroadcastCmd)
	showCmd.AddCommand(showMempoolSizeCmd)
	showCmd.AddCommand(showMempoolMaxTxsBytesCmd)
	showCmd.AddCommand(showMempoolCacheSizeCmd)
	showCmd.AddCommand(showMempoolKeepInvalidTxsInCacheCmd)
	showCmd.AddCommand(showMempoolMaxTxBytesCmd)
	showCmd.AddCommand(showMempoolMaxBatchBytesCmd)

	showCmd.AddCommand(showStatesyncEnableCmd)
	showCmd.AddCommand(showStatesyncRPCServersCmd)
	showCmd.AddCommand(showStatesyncTrustHeightCmd)
	showCmd.AddCommand(showStatesyncTrustHashCmd)
	showCmd.AddCommand(showStatesyncTrustPeriodCmd)
	showCmd.AddCommand(showStatesyncDiscoveryTimeCmd)
	showCmd.AddCommand(showStatesyncTempDirCmd)
	showCmd.AddCommand(showStatesyncChunkRequestTimeoutCmd)
	showCmd.AddCommand(showStatesyncChunkFetchersCmd)

	showCmd.AddCommand(showBlocksyncVersionCmd)

	showCmd.AddCommand(showConsensusWALFileCmd)
	showCmd.AddCommand(showConsensusTimeoutProposeCmd)
	showCmd.AddCommand(showConsensusTimeoutProposeDeltaCmd)
	showCmd.AddCommand(showConsensusTimeoutPrevoteCmd)
	showCmd.AddCommand(showConsensusTimeoutPrevoteDeltaCmd)
	showCmd.AddCommand(showConsensusTimeoutPrecommitCmd)
	showCmd.AddCommand(showConsensusTimeoutPrecommitDeltaCmd)
	showCmd.AddCommand(showConsensusTimeoutCommitCmd)
	showCmd.AddCommand(showConsensusDoubleSignCheckHeightCmd)
	showCmd.AddCommand(showConsensusSkipTimeoutCommitCmd)
	showCmd.AddCommand(showConsensusCreateEmptyBlocksCmd)
	showCmd.AddCommand(showConsensusCreateEmptyBlocksIntervalCmd)
	showCmd.AddCommand(showConsensusPeerGossipSleepDurationCmd)
	showCmd.AddCommand(showConsensusPeerQueryMaj23SleepDurationCmd)

	showCmd.AddCommand(showStorageDiscardABCIResponsesCmd)

	showCmd.AddCommand(showTxIndexIndexerCmd)
	showCmd.AddCommand(showTxIndexPSQLConnCmd)

	showCmd.AddCommand(showInstrumentationPrometheusCmd)
	showCmd.AddCommand(showInstrumentationPrometheusListenAddrCmd)
	showCmd.AddCommand(showInstrumentationMaxOpenConnectionsCmd)
	showCmd.AddCommand(showInstrumentationNamespaceCmd)
}

// runSnapshotInterval executes the 'show snapshot-interval' subcommand
func runSnapshotInterval(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("snapshot-interval: %d\n", config.SnapshotInterval)
	return nil
}

// runSnapshotKeepRecent executes the 'show snapshot-keep-recent' subcommand
func runSnapshotKeepRecent(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("snapshot-keep-recent: %d\n", config.SnapshotKeepRecent)
	return nil
}

// runMinRetainBlocks executes the 'show min-retain-blocks' subcommand
func runMinRetainBlocks(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("min-retain-blocks: %d\n", config.MinRetainBlocks)
	return nil
}

// runPruningMode executes the 'show pruning-mode' subcommand
func runPruningMode(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("pruning-mode: %s\n", config.Pruning)
	return nil
}

// runAppDBBackend executes the 'show app-db-backend' subcommand
func runAppDBBackend(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("app-db-backend: %s\n", config.AppDBBackend)
	return nil
}

// runEVMBuildDelay executes the 'show evm-build-delay' subcommand
func runEVMBuildDelay(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("evm-build-delay: %s\n", config.EVMBuildDelay)
	return nil
}

// runEVMBuildOptimistic executes the 'show evm-build-optimistic' subcommand
func runEVMBuildOptimistic(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("evm-build-optimistic: %t\n", config.EVMBuildOptimistic)
	return nil
}

// runAPIEnable executes the 'show api-enable' subcommand
func runAPIEnable(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("api-enable: %t\n", config.APIEnable)
	return nil
}

// runAPIAddress executes the 'show api-address' subcommand
func runAPIAddress(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("api-address: %s\n", config.APIAddress)
	return nil
}

// runEnabledUnsafeCORS executes the 'show enabled-unsafe-cors' subcommand
func runEnabledUnsafeCORS(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("enabled-unsafe-cors: %t\n", config.EnabledUnsafeCORS)
	return nil
}

// runLogLevel executes the 'show log-level' subcommand
func runLogLevel(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("log-level: %s\n", config.Log.Level)
	return nil
}

// runLogFormat executes the 'show log-format' subcommand
func runLogFormat(cmd *cobra.Command, args []string) error {
	config, err := loadStoryConfig()
	if err != nil {
		return err
	}

	fmt.Printf("log-format: %s\n", config.Log.Format)
	return nil
}

// Alt Komutların RunE Fonksiyonları (config.toml)

// runProxyApp executes the 'show proxy-app' subcommand
func runProxyApp(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("proxy_app: %s\n", config.ProxyApp)
	return nil
}

// runMoniker executes the 'show moniker' subcommand
func runMoniker(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("moniker: %s\n", config.Moniker)
	return nil
}

// runDBDir executes the 'show db-dir' subcommand
func runDBDir(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("db_dir: %s\n", config.DBDir)
	return nil
}

// runGenesisFile executes the 'show genesis-file' subcommand
func runGenesisFile(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("genesis_file: %s\n", config.GenesisFile)
	return nil
}

// runPrivValidatorKeyFile executes the 'show priv-validator-key-file' subcommand
func runPrivValidatorKeyFile(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("priv_validator_key_file: %s\n", config.PrivValidatorKeyFile)
	return nil
}

// runPrivValidatorStateFile executes the 'show priv-validator-state-file' subcommand
func runPrivValidatorStateFile(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("priv_validator_state_file: %s\n", config.PrivValidatorStateFile)
	return nil
}

// runNodeKeyFile executes the 'show node-key-file' subcommand
func runNodeKeyFile(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("node_key_file: %s\n", config.NodeKeyFile)
	return nil
}

// runABCI executes the 'show abci' subcommand
func runABCI(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("abci: %s\n", config.ABCI)
	return nil
}

// runFilterPeers executes the 'show filter-peers' subcommand
func runFilterPeers(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("filter_peers: %t\n", config.FilterPeers)
	return nil
}

// runInstrumentationPrometheus executes the 'show instrumentation-prometheus' subcommand
func runInstrumentationPrometheus(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("instrumentation.prometheus: %t\n", config.Instrumentation.Prometheus)
	return nil
}

// runInstrumentationPrometheusListenAddr executes the 'show instrumentation-prometheus-listen-addr' subcommand
func runInstrumentationPrometheusListenAddr(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("instrumentation.prometheus_listen_addr: %s\n", config.Instrumentation.PrometheusListenAddr)
	return nil
}

// runInstrumentationMaxOpenConnections executes the 'show instrumentation-max-open-connections' subcommand
func runInstrumentationMaxOpenConnections(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("instrumentation.max_open_connections: %d\n", config.Instrumentation.MaxOpenConnections)
	return nil
}

// runInstrumentationNamespace executes the 'show instrumentation-namespace' subcommand
func runInstrumentationNamespace(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("instrumentation.namespace: %s\n", config.Instrumentation.Namespace)
	return nil
}

// loadStoryConfig reads and parses the story.toml file
func loadStoryConfig() (*StoryConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home directory: %v\n", err)
		return nil, err
	}

	storyTomlPath := filepath.Join(homeDir, ".story", "story", "config", "story.toml")

	// Check if story.toml exists
	if _, err := os.Stat(storyTomlPath); os.IsNotExist(err) {
		fmt.Printf("story.toml not found at %s\n", storyTomlPath)
		return nil, fmt.Errorf("story.toml not found")
	}

	var config StoryConfig
	if _, err := toml.DecodeFile(storyTomlPath, &config); err != nil {
		fmt.Printf("Failed to parse story.toml: %v\n", err)
		return nil, err
	}

	return &config, nil
}

// runRPCLaddr executes the 'show rpc-laddr' subcommand
func runRPCLaddr(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.laddr: %s\n", config.RPC.Laddr)
	return nil
}

// runRPCGRPCLaddr executes the 'show rpc-grpc-laddr' subcommand
func runRPCGRPCLaddr(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.grpc_laddr: %s\n", config.RPC.GRPCLaddr)
	return nil
}

// runRPCMaxOpenConnections executes the 'show rpc-grpc-max-open-connections' subcommand
func runRPCMaxOpenConnections(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.grpc_max_open_connections: %d\n", config.RPC.GRPCMaxOpenConnections)
	return nil
}

// runRPCUnsafe executes the 'show rpc-unsafe' subcommand
func runRPCUnsafe(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.unsafe: %t\n", config.RPC.Unsafe)
	return nil
}

// runRPCMaxSubscriptionClients executes the 'show rpc-max-subscription-clients' subcommand
func runRPCMaxSubscriptionClients(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.max_subscription_clients: %d\n", config.RPC.MaxSubscriptionClients)
	return nil
}

// runRPCMaxSubscriptionsPerClient executes the 'show rpc-max-subscriptions-per-client' subcommand
func runRPCMaxSubscriptionsPerClient(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.max_subscriptions_per_client: %d\n", config.RPC.MaxSubscriptionsPerClient)
	return nil
}

// runRPCExperimentalSubscriptionBufferSize executes the 'show rpc-experimental-subscription-buffer-size' subcommand
func runRPCExperimentalSubscriptionBufferSize(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.experimental_subscription_buffer_size: %d\n", config.RPC.ExperimentalSubscriptionBufferSize)
	return nil
}

// runRPCExperimentalWebsocketWriteBufferSize executes the 'show rpc-experimental-websocket-write-buffer-size' subcommand
func runRPCExperimentalWebsocketWriteBufferSize(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.experimental_websocket_write_buffer_size: %d\n", config.RPC.ExperimentalWebsocketWriteBufferSize)
	return nil
}

// runRPCExperimentalCloseOnSlowClient executes the 'show rpc-experimental-close-on-slow-client' subcommand
func runRPCExperimentalCloseOnSlowClient(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.experimental_close_on_slow_client: %t\n", config.RPC.ExperimentalCloseOnSlowClient)
	return nil
}

// runRPCTimeoutBroadcastTxCommit executes the 'show rpc-timeout-broadcast-tx-commit' subcommand
func runRPCTimeoutBroadcastTxCommit(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.timeout_broadcast_tx_commit: %s\n", config.RPC.TimeoutBroadcastTxCommit)
	return nil
}

// runRPCMaxRequestBatchSize executes the 'show rpc-max-request-batch-size' subcommand
func runRPCMaxRequestBatchSize(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.max_request_batch_size: %d\n", config.RPC.MaxRequestBatchSize)
	return nil
}

// runRPCMaxBodyBytes executes the 'show rpc-max-body-bytes' subcommand
func runRPCMaxBodyBytes(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.max_body_bytes: %d\n", config.RPC.MaxBodyBytes)
	return nil
}

// runRPCMaxHeaderBytes executes the 'show rpc-max-header-bytes' subcommand
func runRPCMaxHeaderBytes(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.max_header_bytes: %d\n", config.RPC.MaxHeaderBytes)
	return nil
}

// runRPCPPROFLaddr executes the 'show rpc-pprof-laddr' subcommand
func runRPCPPROFLaddr(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("rpc.pprof_laddr: %s\n", config.RPC.PPROFLaddr)
	return nil
}

// runP2PLaddr executes the 'show p2p-laddr' subcommand
func runP2PLaddr(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.laddr: %s\n", config.P2P.Laddr)
	return nil
}

// runP2PExternalAddress executes the 'show p2p-external-address' subcommand
func runP2PExternalAddress(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.external_address: %s\n", config.P2P.ExternalAddress)
	return nil
}

// runP2PSeeds executes the 'show p2p-seeds' subcommand
func runP2PSeeds(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.seeds: %s\n", config.P2P.Seeds)
	return nil
}

// runP2PPersistentPeers executes the 'show p2p-persistent-peers' subcommand
func runP2PPersistentPeers(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.persistent_peers: %s\n", config.P2P.PersistentPeers)
	return nil
}

// runP2PAddrBookFile executes the 'show p2p-addr-book-file' subcommand
func runP2PAddrBookFile(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.addr_book_file: %s\n", config.P2P.AddrBookFile)
	return nil
}

// runP2PAddrBookStrict executes the 'show p2p-addr-book-strict' subcommand
func runP2PAddrBookStrict(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.addr_book_strict: %t\n", config.P2P.AddrBookStrict)
	return nil
}

// runP2PMaxNumInboundPeers executes the 'show p2p-max-num-inbound-peers' subcommand
func runP2PMaxNumInboundPeers(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.max_num_inbound_peers: %d\n", config.P2P.MaxNumInboundPeers)
	return nil
}

// runP2PMaxNumOutboundPeers executes the 'show p2p-max-num-outbound-peers' subcommand
func runP2PMaxNumOutboundPeers(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.max_num_outbound_peers: %d\n", config.P2P.MaxNumOutboundPeers)
	return nil
}

// runP2PUnconditionalPeerIDs executes the 'show p2p-unconditional-peer-ids' subcommand
func runP2PUnconditionalPeerIDs(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.unconditional_peer_ids: %s\n", config.P2P.UnconditionalPeerIDs)
	return nil
}

// runP2PPersistentPeersMaxDialPeriod executes the 'show p2p-persistent-peers-max-dial-period' subcommand
func runP2PPersistentPeersMaxDialPeriod(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.persistent_peers_max_dial_period: %s\n", config.P2P.PersistentPeersMaxDialPeriod)
	return nil
}

// runP2PFlushThrottleTimeout executes the 'show p2p-flush-throttle-timeout' subcommand
func runP2PFlushThrottleTimeout(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.flush_throttle_timeout: %s\n", config.P2P.FlushThrottleTimeout)
	return nil
}

// runP2PMaxPacketMsgPayloadSize executes the 'show p2p-max-packet-msg-payload-size' subcommand
func runP2PMaxPacketMsgPayloadSize(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.max_packet_msg_payload_size: %d\n", config.P2P.MaxPacketMsgPayloadSize)
	return nil
}

// runP2PSendRate executes the 'show p2p-send-rate' subcommand
func runP2PSendRate(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.send_rate: %d\n", config.P2P.SendRate)
	return nil
}

// runP2PRecvRate executes the 'show p2p-recv-rate' subcommand
func runP2PRecvRate(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.recv_rate: %d\n", config.P2P.RecvRate)
	return nil
}

// runP2PPEX executes the 'show p2p-pex' subcommand
func runP2PPEX(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.pex: %t\n", config.P2P.PEX)
	return nil
}

// runP2PSeedMode executes the 'show p2p-seed-mode' subcommand
func runP2PSeedMode(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.seed_mode: %t\n", config.P2P.SeedMode)
	return nil
}

// runP2PAllowDuplicateIP executes the 'show p2p-allow-duplicate-ip' subcommand
func runP2PAllowDuplicateIP(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.allow_duplicate_ip: %t\n", config.P2P.AllowDuplicateIP)
	return nil
}

// runP2PHandshakeTimeout executes the 'show p2p-handshake-timeout' subcommand
func runP2PHandshakeTimeout(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.handshake_timeout: %s\n", config.P2P.HandshakeTimeout)
	return nil
}

// runP2PDialTimeout executes the 'show p2p-dial-timeout' subcommand
func runP2PDialTimeout(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("p2p.dial_timeout: %s\n", config.P2P.DialTimeout)
	return nil
}

// runMempoolType executes the 'show mempool-type' subcommand
func runMempoolType(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.type: %s\n", config.Mempool.Type)
	return nil
}

// runMempoolRecheck executes the 'show mempool-recheck' subcommand
func runMempoolRecheck(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.recheck: %t\n", config.Mempool.Recheck)
	return nil
}

// runMempoolRecheckTimeout executes the 'show mempool-recheck-timeout' subcommand
func runMempoolRecheckTimeout(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.recheck_timeout: %s\n", config.Mempool.RecheckTimeout)
	return nil
}

// runMempoolBroadcast executes the 'show mempool-broadcast' subcommand
func runMempoolBroadcast(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.broadcast: %t\n", config.Mempool.Broadcast)
	return nil
}

// runMempoolSize executes the 'show mempool-size' subcommand
func runMempoolSize(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.size: %d\n", config.Mempool.Size)
	return nil
}

// runMempoolMaxTxsBytes executes the 'show mempool-max-txs-bytes' subcommand
func runMempoolMaxTxsBytes(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.max_txs_bytes: %d\n", config.Mempool.MaxTxsBytes)
	return nil
}

// runMempoolCacheSize executes the 'show mempool-cache-size' subcommand
func runMempoolCacheSize(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.cache_size: %d\n", config.Mempool.CacheSize)
	return nil
}

// runMempoolKeepInvalidTxsInCache executes the 'show mempool-keep-invalid-txs-in-cache' subcommand
func runMempoolKeepInvalidTxsInCache(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.keep-invalid-txs-in-cache: %t\n", config.Mempool.KeepInvalidTxsInCache)
	return nil
}

// runMempoolMaxTxBytes executes the 'show mempool-max-tx-bytes' subcommand
func runMempoolMaxTxBytes(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.max_tx_bytes: %d\n", config.Mempool.MaxTxBytes)
	return nil
}

// runMempoolMaxBatchBytes executes the 'show mempool-max-batch-bytes' subcommand
func runMempoolMaxBatchBytes(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("mempool.max_batch_bytes: %d\n", config.Mempool.MaxBatchBytes)
	return nil
}

// runStatesyncEnable executes the 'show statesync-enable' subcommand
func runStatesyncEnable(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.enable: %t\n", config.Statesync.Enable)
	return nil
}

// runStatesyncRPCServers executes the 'show statesync-rpc-servers' subcommand
func runStatesyncRPCServers(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.rpc_servers: %s\n", config.Statesync.RPCServers)
	return nil
}

// runStatesyncTrustHeight executes the 'show statesync-trust-height' subcommand
func runStatesyncTrustHeight(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.trust_height: %d\n", config.Statesync.TrustHeight)
	return nil
}

// runStatesyncTrustHash executes the 'show statesync-trust-hash' subcommand
func runStatesyncTrustHash(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.trust_hash: %s\n", config.Statesync.TrustHash)
	return nil
}

// runStatesyncTrustPeriod executes the 'show statesync-trust-period' subcommand
func runStatesyncTrustPeriod(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.trust_period: %s\n", config.Statesync.TrustPeriod)
	return nil
}

// runStatesyncDiscoveryTime executes the 'show statesync-discovery-time' subcommand
func runStatesyncDiscoveryTime(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.discovery_time: %s\n", config.Statesync.DiscoveryTime)
	return nil
}

// runStatesyncTempDir executes the 'show statesync-temp-dir' subcommand
func runStatesyncTempDir(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.temp_dir: %s\n", config.Statesync.TempDir)
	return nil
}

// runStatesyncChunkRequestTimeout executes the 'show statesync-chunk-request-timeout' subcommand
func runStatesyncChunkRequestTimeout(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.chunk_request_timeout: %s\n", config.Statesync.ChunkRequestTimeout)
	return nil
}

// runStatesyncChunkFetchers executes the 'show statesync-chunk-fetchers' subcommand
func runStatesyncChunkFetchers(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("statesync.chunk_fetchers: %s\n", config.Statesync.ChunkFetchers)
	return nil
}

// runBlocksyncVersion executes the 'show blocksync-version' subcommand
func runBlocksyncVersion(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("blocksync.version: %s\n", config.Blocksync.Version)
	return nil
}

// runConsensusWALFile executes the 'show consensus-wal-file' subcommand
func runConsensusWALFile(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.wal_file: %s\n", config.Consensus.WALFile)
	return nil
}

// runConsensusTimeoutPropose executes the 'show consensus-timeout-propose' subcommand
func runConsensusTimeoutPropose(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.timeout_propose: %s\n", config.Consensus.TimeoutPropose)
	return nil
}

// runConsensusTimeoutProposeDelta executes the 'show consensus-timeout-propose-delta' subcommand
func runConsensusTimeoutProposeDelta(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.timeout_propose_delta: %s\n", config.Consensus.TimeoutProposeDelta)
	return nil
}

// runConsensusTimeoutPrevote executes the 'show consensus-timeout-prevote' subcommand
func runConsensusTimeoutPrevote(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.timeout_prevote: %s\n", config.Consensus.TimeoutPrevote)
	return nil
}

// runConsensusTimeoutPrevoteDelta executes the 'show consensus-timeout-prevote-delta' subcommand
func runConsensusTimeoutPrevoteDelta(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.timeout_prevote_delta: %s\n", config.Consensus.TimeoutPrevoteDelta)
	return nil
}

// runConsensusTimeoutPrecommit executes the 'show consensus-timeout-precommit' subcommand
func runConsensusTimeoutPrecommit(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.timeout_precommit: %s\n", config.Consensus.TimeoutPrecommit)
	return nil
}

// runConsensusTimeoutPrecommitDelta executes the 'show consensus-timeout-precommit-delta' subcommand
func runConsensusTimeoutPrecommitDelta(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.timeout_precommit_delta: %s\n", config.Consensus.TimeoutPrecommitDelta)
	return nil
}

// runConsensusTimeoutCommit executes the 'show consensus-timeout-commit' subcommand
func runConsensusTimeoutCommit(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.timeout_commit: %s\n", config.Consensus.TimeoutCommit)
	return nil
}

// runConsensusDoubleSignCheckHeight executes the 'show consensus-double-sign-check-height' subcommand
func runConsensusDoubleSignCheckHeight(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.double_sign_check_height: %d\n", config.Consensus.DoubleSignCheckHeight)
	return nil
}

// runConsensusSkipTimeoutCommit executes the 'show consensus-skip-timeout-commit' subcommand
func runConsensusSkipTimeoutCommit(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.skip_timeout_commit: %t\n", config.Consensus.SkipTimeoutCommit)
	return nil
}

// runConsensusCreateEmptyBlocks executes the 'show consensus-create-empty-blocks' subcommand
func runConsensusCreateEmptyBlocks(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.create_empty_blocks: %t\n", config.Consensus.CreateEmptyBlocks)
	return nil
}

// runConsensusCreateEmptyBlocksInterval executes the 'show consensus-create-empty-blocks-interval' subcommand
func runConsensusCreateEmptyBlocksInterval(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.create_empty_blocks_interval: %s\n", config.Consensus.CreateEmptyBlocksInterval)
	return nil
}

// runConsensusPeerGossipSleepDuration executes the 'show consensus-peer-gossip-sleep-duration' subcommand
func runConsensusPeerGossipSleepDuration(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.peer_gossip_sleep_duration: %s\n", config.Consensus.PeerGossipSleepDuration)
	return nil
}

// runConsensusPeerQueryMaj23SleepDuration executes the 'show consensus-peer-query-maj23-sleep-duration' subcommand
func runConsensusPeerQueryMaj23SleepDuration(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("consensus.peer_query_maj23_sleep_duration: %s\n", config.Consensus.PeerQueryMaj23SleepDuration)
	return nil
}

// runStorageDiscardABCIResponses executes the 'show storage-discard-abci-responses' subcommand
func runStorageDiscardABCIResponses(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("storage.discard_abci_responses: %t\n", config.Storage.DiscardABCIResponses)
	return nil
}

// runTxIndexIndexer executes the 'show tx-index-indexer' subcommand
func runTxIndexIndexer(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("tx_index.indexer: %s\n", config.TxIndex.Indexer)
	return nil
}

// runTxIndexPSQLConn executes the 'show tx-index-psql-conn' subcommand
func runTxIndexPSQLConn(cmd *cobra.Command, args []string) error {
	config, err := loadConfigConfig()
	if err != nil {
		return err
	}

	fmt.Printf("tx_index.psql-conn: %s\n", config.TxIndex.PSQLConn)
	return nil
}

// loadConfigConfig reads and parses the config.toml file
func loadConfigConfig() (*ConfigConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Failed to get user home directory: %v\n", err)
		return nil, err
	}

	configTomlPath := filepath.Join(homeDir, ".story", "story", "config", "config.toml")

	// Check if config.toml exists
	if _, err := os.Stat(configTomlPath); os.IsNotExist(err) {
		fmt.Printf("config.toml not found at %s\n", configTomlPath)
		return nil, fmt.Errorf("config.toml not found")
	}

	var config ConfigConfig
	if _, err := toml.DecodeFile(configTomlPath, &config); err != nil {
		fmt.Printf("Failed to parse config.toml: %v\n", err)
		return nil, err
	}

	return &config, nil
}

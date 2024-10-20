package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

// setCmd represents the set command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set configuration values in Story and CometBFT services",
	Long:  `Provides subcommands to set specific configuration values in story.toml and config.toml.`,
}

// Structs representing the configuration sections
type SetStoryConfig struct {
	Pruning            string       `toml:"pruning"`
	SnapshotInterval   int          `toml:"snapshot-interval"`
	SnapshotKeepRecent int          `toml:"snapshot-keep-recent"`
	MinRetainBlocks    int          `toml:"min_retain_blocks"`
	AppDBBackend       string       `toml:"app_db_backend"`
	EVMBuildDelay      string       `toml:"evm_build_delay"`
	EVMBuildOptimistic bool         `toml:"evm_build_optimistic"`
	APIEnable          bool         `toml:"api-enable"`
	APIAddress         string       `toml:"api-address"`
	EnabledUnsafeCORS  bool         `toml:"enabled-unsafe-cors"`
	Log                SetLogConfig `toml:"log"`
}

type SetLogConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
}

type SetConfigConfig struct {
	Version                string                   `toml:"version"`
	ProxyApp               string                   `toml:"proxy_app"`
	Moniker                string                   `toml:"moniker"`
	DBBackend              string                   `toml:"db_backend"`
	DBDir                  string                   `toml:"db_dir"`
	LogLevel               string                   `toml:"log_level"`
	LogFormat              string                   `toml:"log_format"`
	GenesisFile            string                   `toml:"genesis_file"`
	PrivValidatorKeyFile   string                   `toml:"priv_validator_key_file"`
	PrivValidatorStateFile string                   `toml:"priv_validator_state_file"`
	NodeKeyFile            string                   `toml:"node_key_file"`
	ABCI                   string                   `toml:"abci"`
	FilterPeers            bool                     `toml:"filter_peers"`
	RPC                    SetRPCConfig             `toml:"rpc"`
	P2P                    SetP2PConfig             `toml:"p2p"`
	Mempool                SetMempoolConfig         `toml:"mempool"`
	Statesync              SetStatesyncConfig       `toml:"statesync"`
	Blocksync              SetBlocksyncConfig       `toml:"blocksync"`
	Consensus              SetConsensusConfig       `toml:"consensus"`
	Storage                SetStorageConfig         `toml:"storage"`
	TxIndex                SetTxIndexConfig         `toml:"tx_index"`
	Instrumentation        SetInstrumentationConfig `toml:"instrumentation"`
}

type SetRPCConfig struct {
	Laddr                                string `toml:"laddr"`
	GRPCLaddr                            string `toml:"grpc_laddr"`
	GRPCMaxOpenConnections               int    `toml:"grpc_max_open_connections"`
	Unsafe                               bool   `toml:"unsafe"`
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

type SetP2PConfig struct {
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

type SetMempoolConfig struct {
	Type                  string `toml:"type"`
	Recheck               bool   `toml:"recheck"`
	RecheckTimeout        string `toml:"recheck_timeout"`
	Broadcast             bool   `toml:"broadcast"`
	Size                  int    `toml:"size"`
	MaxTxsBytes           int    `toml:"max_txs_bytes"`
	CacheSize             int    `toml:"cache_size"`
	KeepInvalidTxsInCache bool   `toml:"keep_invalid_txs_in_cache"`
	MaxTxBytes            int    `toml:"max_tx_bytes"`
	MaxBatchBytes         int    `toml:"max_batch_bytes"`
}

type SetStatesyncConfig struct {
	Enable              bool   `toml:"enable"`
	RPCServers          string `toml:"rpc_servers"`
	TrustHeight         int    `toml:"trust_height"`
	TrustHash           string `toml:"trust_hash"`
	TrustPeriod         string `toml:"trust_period"`
	DiscoveryTime       string `toml:"discovery_time"`
	TempDir             string `toml:"temp_dir"`
	ChunkRequestTimeout string `toml:"chunk_request_timeout"`
	ChunkFetchers       int    `toml:"chunk_fetchers"`
}

type SetBlocksyncConfig struct {
	Version string `toml:"version"`
}

type SetConsensusConfig struct {
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

type SetStorageConfig struct {
	DiscardABCIResponses bool `toml:"discard_abci_responses"`
}

type SetTxIndexConfig struct {
	Indexer  string `toml:"indexer"`
	PSQLConn string `toml:"psql_conn"`
}

type SetInstrumentationConfig struct {
	Prometheus           bool   `toml:"prometheus"`
	PrometheusListenAddr string `toml:"prometheus_listen_addr"`
	MaxOpenConnections   int    `toml:"max_open_connections"`
	Namespace            string `toml:"namespace"`
}

func init() {
	// Add 'set' command to root
	rootCmd.AddCommand(setCmd)

	// Add set subcommands for story.toml
	setCmd.AddCommand(setPruningModeCmd)
	setCmd.AddCommand(setSnapshotIntervalCmd)
	setCmd.AddCommand(setSnapshotKeepRecentCmd)
	setCmd.AddCommand(setMinRetainBlocksCmd)
	setCmd.AddCommand(setAppDBBackendCmd)
	setCmd.AddCommand(setEVMBuildDelayCmd)
	setCmd.AddCommand(setEVMBuildOptimisticCmd)
	setCmd.AddCommand(setAPIEnableCmd)
	setCmd.AddCommand(setAPIAddressCmd)
	setCmd.AddCommand(setEnabledUnsafeCORSCmd)
	setCmd.AddCommand(setLogLevelCmd)
	setCmd.AddCommand(setLogFormatCmd)

	// Add set subcommands for config.toml
	setCmd.AddCommand(setConfigVersionCmd)
	setCmd.AddCommand(setProxyAppCmd)
	setCmd.AddCommand(setMonikerCmd)
	setCmd.AddCommand(setDBBackendCmd)
	setCmd.AddCommand(setDBDirCmd)
	setCmd.AddCommand(setLogLevelConfigCmd)
	setCmd.AddCommand(setLogFormatConfigCmd)
	setCmd.AddCommand(setGenesisFileCmd)
	setCmd.AddCommand(setPrivValidatorKeyFileCmd)
	setCmd.AddCommand(setPrivValidatorStateFileCmd)
	setCmd.AddCommand(setNodeKeyFileCmd)
	setCmd.AddCommand(setABCICommand)
	setCmd.AddCommand(setFilterPeersCmd)

	// Add set subcommands for [rpc] section
	setCmd.AddCommand(setRPCLaddrCmd)
	setCmd.AddCommand(setRPCGRPCLaddrCmd)
	setCmd.AddCommand(setRPCMaxOpenConnectionsCmd)
	setCmd.AddCommand(setRPCUnsafeCmd)
	setCmd.AddCommand(setRPCMaxSubscriptionClientsCmd)
	setCmd.AddCommand(setRPCMaxSubscriptionsPerClientCmd)
	setCmd.AddCommand(setRPCExperimentalSubscriptionBufferSizeCmd)
	setCmd.AddCommand(setRPCExperimentalWebsocketWriteBufferSizeCmd)
	setCmd.AddCommand(setRPCExperimentalCloseOnSlowClientCmd)
	setCmd.AddCommand(setRPCTimeoutBroadcastTxCommitCmd)
	setCmd.AddCommand(setRPCMaxRequestBatchSizeCmd)
	setCmd.AddCommand(setRPCMaxBodyBytesCmd)
	setCmd.AddCommand(setRPCMaxHeaderBytesCmd)
	setCmd.AddCommand(setRPCPPROFLaddrCmd)

	// Add set subcommands for [p2p] section
	setCmd.AddCommand(setP2PLaddrCmd)
	setCmd.AddCommand(setP2PExternalAddressCmd)
	setCmd.AddCommand(setP2PSeedsCmd)
	setCmd.AddCommand(setP2PPersistentPeersCmd)
	setCmd.AddCommand(setP2PAddrBookFileCmd)
	setCmd.AddCommand(setP2PAddrBookStrictCmd)
	setCmd.AddCommand(setP2PMaxNumInboundPeersCmd)
	setCmd.AddCommand(setP2PMaxNumOutboundPeersCmd)
	setCmd.AddCommand(setP2PUnconditionalPeerIDsCmd)
	setCmd.AddCommand(setP2PPersistentPeersMaxDialPeriodCmd)
	setCmd.AddCommand(setP2PFlushThrottleTimeoutCmd)
	setCmd.AddCommand(setP2PMaxPacketMsgPayloadSizeCmd)
	setCmd.AddCommand(setP2PSendRateCmd)
	setCmd.AddCommand(setP2PRecvRateCmd)
	setCmd.AddCommand(setP2PPEXCmd)
	setCmd.AddCommand(setP2PSeedModeCmd)
	setCmd.AddCommand(setP2PAllowDuplicateIPCmd)
	setCmd.AddCommand(setP2PHandshakeTimeoutCmd)
	setCmd.AddCommand(setP2PDialTimeoutCmd)

	// Add set subcommands for [mempool] section
	setCmd.AddCommand(setMempoolTypeCmd)
	setCmd.AddCommand(setMempoolRecheckCmd)
	setCmd.AddCommand(setMempoolRecheckTimeoutCmd)
	setCmd.AddCommand(setMempoolBroadcastCmd)
	setCmd.AddCommand(setMempoolSizeCmd)
	setCmd.AddCommand(setMempoolMaxTxsBytesCmd)
	setCmd.AddCommand(setMempoolCacheSizeCmd)
	setCmd.AddCommand(setMempoolKeepInvalidTxsInCacheCmd)
	setCmd.AddCommand(setMempoolMaxTxBytesCmd)
	setCmd.AddCommand(setMempoolMaxBatchBytesCmd)

	// Add set subcommands for [statesync] section
	setCmd.AddCommand(setStatesyncEnableCmd)
	setCmd.AddCommand(setStatesyncRPCServersCmd)
	setCmd.AddCommand(setStatesyncTrustHeightCmd)
	setCmd.AddCommand(setStatesyncTrustHashCmd)
	setCmd.AddCommand(setStatesyncTrustPeriodCmd)
	setCmd.AddCommand(setStatesyncDiscoveryTimeCmd)
	setCmd.AddCommand(setStatesyncTempDirCmd)
	setCmd.AddCommand(setStatesyncChunkRequestTimeoutCmd)
	setCmd.AddCommand(setStatesyncChunkFetchersCmd)

	// Add set subcommands for [blocksync] section
	setCmd.AddCommand(setBlocksyncVersionCmd)

	// Add set subcommands for [consensus] section
	setCmd.AddCommand(setConsensusWALFileCmd)
	setCmd.AddCommand(setConsensusTimeoutProposeCmd)
	setCmd.AddCommand(setConsensusTimeoutProposeDeltaCmd)
	setCmd.AddCommand(setConsensusTimeoutPrevoteCmd)
	setCmd.AddCommand(setConsensusTimeoutPrevoteDeltaCmd)
	setCmd.AddCommand(setConsensusTimeoutPrecommitCmd)
	setCmd.AddCommand(setConsensusTimeoutPrecommitDeltaCmd)
	setCmd.AddCommand(setConsensusTimeoutCommitCmd)
	setCmd.AddCommand(setConsensusDoubleSignCheckHeightCmd)
	setCmd.AddCommand(setConsensusSkipTimeoutCommitCmd)
	setCmd.AddCommand(setConsensusCreateEmptyBlocksCmd)
	setCmd.AddCommand(setConsensusCreateEmptyBlocksIntervalCmd)
	setCmd.AddCommand(setConsensusPeerGossipSleepDurationCmd)
	setCmd.AddCommand(setConsensusPeerQueryMaj23SleepDurationCmd)

	// Add set subcommands for [storage] section
	setCmd.AddCommand(setStorageDiscardABCIResponsesCmd)

	// Add set subcommands for [tx_index] section
	setCmd.AddCommand(setTxIndexIndexerCmd)
	setCmd.AddCommand(setTxIndexPSQLConnCmd)

	// Add set subcommands for [instrumentation] section
	setCmd.AddCommand(setInstrumentationPrometheusCmd)
	setCmd.AddCommand(setInstrumentationPrometheusListenAddrCmd)
	setCmd.AddCommand(setInstrumentationMaxOpenConnectionsCmd)
	setCmd.AddCommand(setInstrumentationNamespaceCmd)
}

var setPruningModeCmd = &cobra.Command{
	Use:   "pruning-mode [value]",
	Short: "Set the pruning mode value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetPruningMode,
}

// runSetPruningMode sets the pruning mode in story.toml
func runSetPruningMode(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	// Validate the newValue against allowed pruning modes
	allowedValues := []string{"default", "nothing", "everything", "custom"}
	isValid := false
	for _, v := range allowedValues {
		if newValue == v {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid pruning mode: %s. Allowed values are: %v", newValue, allowedValues)
	}

	// Load the current configuration
	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	// Update the pruning mode
	config.Pruning = newValue

	// Save the updated configuration
	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("pruning-mode set to: %s", newValue))
	return nil
}

// setSnapshotIntervalCmd sets the snapshot_interval in story.toml
var setSnapshotIntervalCmd = &cobra.Command{
	Use:   "snapshot-interval [value]",
	Short: "Set the snapshot-interval value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetSnapshotInterval,
}

func runSetSnapshotInterval(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid snapshot-interval value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.SnapshotInterval = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("snapshot-interval set to: %d", newValue))
	return nil
}

// setSnapshotKeepRecentCmd sets the snapshot_keep_recent in story.toml
var setSnapshotKeepRecentCmd = &cobra.Command{
	Use:   "snapshot-keep-recent [value]",
	Short: "Set the snapshot-keep-recent value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetSnapshotKeepRecent,
}

func runSetSnapshotKeepRecent(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid snapshot-keep-recent value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.SnapshotKeepRecent = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("snapshot-keep-recent set to: %d", newValue))
	return nil
}

// setMinRetainBlocksCmd sets the min_retain_blocks in story.toml
var setMinRetainBlocksCmd = &cobra.Command{
	Use:   "min-retain-blocks [value]",
	Short: "Set the min-retain-blocks value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMinRetainBlocks,
}

func runSetMinRetainBlocks(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid min-retain-blocks value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.MinRetainBlocks = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("min-retain-blocks set to: %d", newValue))
	return nil
}

// setAppDBBackendCmd sets the app_db_backend in story.toml
var setAppDBBackendCmd = &cobra.Command{
	Use:   "app-db-backend [value]",
	Short: "Set the app-db-backend value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetAppDBBackend,
}

func runSetAppDBBackend(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.AppDBBackend = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("app-db-backend set to: %s", newValue))
	return nil
}

// setEVMBuildDelayCmd sets the evm_build_delay in story.toml
var setEVMBuildDelayCmd = &cobra.Command{
	Use:   "evm-build-delay [value]",
	Short: "Set the evm-build-delay value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetEVMBuildDelay,
}

func runSetEVMBuildDelay(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.EVMBuildDelay = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("evm-build-delay set to: %s", newValue))
	return nil
}

// setEVMBuildOptimisticCmd sets the evm_build_optimistic in story.toml
var setEVMBuildOptimisticCmd = &cobra.Command{
	Use:   "evm-build-optimistic [value]",
	Short: "Set the evm-build-optimistic value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetEVMBuildOptimistic,
}

func runSetEVMBuildOptimistic(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid evm-build-optimistic value: %s. Must be true or false", input)
	}

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.EVMBuildOptimistic = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("evm-build-optimistic set to: %t", newValue))
	return nil
}

// setAPIEnableCmd sets the api_enable in story.toml
var setAPIEnableCmd = &cobra.Command{
	Use:   "api-enable [value]",
	Short: "Set the api-enable value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetAPIEnable,
}

func runSetAPIEnable(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid api-enable value: %s. Must be true or false", input)
	}

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.APIEnable = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("api-enable set to: %t", newValue))
	return nil
}

// setAPIAddressCmd sets the api_address in story.toml
var setAPIAddressCmd = &cobra.Command{
	Use:   "api-address [value]",
	Short: "Set the api-address value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetAPIAddress,
}

func runSetAPIAddress(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.APIAddress = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("api-address set to: %s", newValue))
	return nil
}

// setEnabledUnsafeCORSCmd sets the enabled_unsafe_cors in story.toml
var setEnabledUnsafeCORSCmd = &cobra.Command{
	Use:   "enabled-unsafe-cors [value]",
	Short: "Set the enabled-unsafe-cors value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetEnabledUnsafeCORS,
}

func runSetEnabledUnsafeCORS(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid enabled-unsafe-cors value: %s. Must be true or false", input)
	}

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	config.EnabledUnsafeCORS = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("enabled-unsafe-cors set to: %t", newValue))
	return nil
}

// setLogLevelCmd sets the log.level in story.toml
var setLogLevelCmd = &cobra.Command{
	Use:   "log-level [value]",
	Short: "Set the log-level value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetLogLevel,
}

func runSetLogLevel(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	// Optional: Validate log levels
	allowedLevels := []string{"debug", "info", "warn", "error"}
	isValid := false
	for _, level := range allowedLevels {
		if newValue == level {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid log level: %s. Allowed values are: %v", newValue, allowedLevels)
	}

	// Update nested [log] section
	config.Log.Level = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("log-level set to: %s", newValue))
	return nil
}

// setLogFormatCmd sets the log.format in story.toml
var setLogFormatCmd = &cobra.Command{
	Use:   "log-format [value]",
	Short: "Set the log-format value in story.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetLogFormat,
}

func runSetLogFormat(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetStoryConfig()
	if err != nil {
		return err
	}

	// Optional: Validate log formats
	allowedFormats := []string{"plain", "json"}
	isValid := false
	for _, format := range allowedFormats {
		if newValue == format {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid log format: %s. Allowed values are: %v", newValue, allowedFormats)
	}

	// Update nested [log] section
	config.Log.Format = newValue

	if err := saveSetStoryConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("log-format set to: %s", newValue))
	return nil
}

// ----------------------- //
// config.toml set commands
// ----------------------- //

// setConfigVersionCmd sets the version in config.toml
var setConfigVersionCmd = &cobra.Command{
	Use:   "config-version [value]",
	Short: "Set the version value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConfigVersion,
}

func runSetConfigVersion(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Version = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("config-version set to: %s", newValue))
	return nil
}

// setProxyAppCmd sets the proxy_app in config.toml
var setProxyAppCmd = &cobra.Command{
	Use:   "proxy-app [value]",
	Short: "Set the proxy_app value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetProxyApp,
}

func runSetProxyApp(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	// Optional: Validate the newValue (e.g., ensure it's a valid address)
	// You can add regex or other validation here

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.ProxyApp = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("proxy_app set to: %s", newValue))
	return nil
}

// setMonikerCmd sets the moniker in config.toml
var setMonikerCmd = &cobra.Command{
	Use:   "moniker [value]",
	Short: "Set the moniker value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMoniker,
}

func runSetMoniker(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Moniker = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("moniker set to: %s", newValue))
	return nil
}

// setDBBackendCmd sets the db_backend in config.toml
var setDBBackendCmd = &cobra.Command{
	Use:   "db-backend [value]",
	Short: "Set the db_backend value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetDBBackend,
}

func runSetDBBackend(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.DBBackend = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("db_backend set to: %s", newValue))
	return nil
}

// setDBDirCmd sets the db_dir in config.toml
var setDBDirCmd = &cobra.Command{
	Use:   "db-dir [value]",
	Short: "Set the db_dir value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetDBDir,
}

func runSetDBDir(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.DBDir = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("db_dir set to: %s", newValue))
	return nil
}

// setLogLevelConfigCmd sets the log_level in config.toml
var setLogLevelConfigCmd = &cobra.Command{
	Use:   "log-level-config [value]",
	Short: "Set the log_level value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetLogLevelConfig,
}

func runSetLogLevelConfig(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.LogLevel = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("log_level set to: %s", newValue))
	return nil
}

// setLogFormatConfigCmd sets the log_format in config.toml
var setLogFormatConfigCmd = &cobra.Command{
	Use:   "log-format-config [value]",
	Short: "Set the log_format value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetLogFormatConfig,
}

func runSetLogFormatConfig(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.LogFormat = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("log_format set to: %s", newValue))
	return nil
}

// setGenesisFileCmd sets the genesis_file in config.toml
var setGenesisFileCmd = &cobra.Command{
	Use:   "genesis-file [value]",
	Short: "Set the genesis_file value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetGenesisFile,
}

func runSetGenesisFile(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.GenesisFile = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("genesis_file set to: %s", newValue))
	return nil
}

// setPrivValidatorKeyFileCmd sets the priv_validator_key_file in config.toml
var setPrivValidatorKeyFileCmd = &cobra.Command{
	Use:   "priv-validator-key-file [value]",
	Short: "Set the priv_validator_key_file value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetPrivValidatorKeyFile,
}

func runSetPrivValidatorKeyFile(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.PrivValidatorKeyFile = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("priv_validator_key_file set to: %s", newValue))
	return nil
}

// setPrivValidatorStateFileCmd sets the priv_validator_state_file in config.toml
var setPrivValidatorStateFileCmd = &cobra.Command{
	Use:   "priv-validator-state-file [value]",
	Short: "Set the priv_validator_state_file value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetPrivValidatorStateFile,
}

func runSetPrivValidatorStateFile(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.PrivValidatorStateFile = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("priv_validator_state_file set to: %s", newValue))
	return nil
}

// setNodeKeyFileCmd sets the node_key_file in config.toml
var setNodeKeyFileCmd = &cobra.Command{
	Use:   "node-key-file [value]",
	Short: "Set the node_key_file value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetNodeKeyFile,
}

func runSetNodeKeyFile(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.NodeKeyFile = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("node_key_file set to: %s", newValue))
	return nil
}

// setABCICommand sets the abci in config.toml
var setABCICommand = &cobra.Command{
	Use:   "abci [value]",
	Short: "Set the abci value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetABCI,
}

func runSetABCI(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.ABCI = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("abci set to: %s", newValue))
	return nil
}

// setFilterPeersCmd sets the filter_peers in config.toml
var setFilterPeersCmd = &cobra.Command{
	Use:   "filter-peers [value]",
	Short: "Set the filter_peers value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetFilterPeers,
}

func runSetFilterPeers(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid filter-peers value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.FilterPeers = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("filter_peers set to: %t", newValue))
	return nil
}

// ----------------------- //
// [rpc] section set commands
// ----------------------- //

// setRPCLaddrCmd sets the rpc.laddr in config.toml
var setRPCLaddrCmd = &cobra.Command{
	Use:   "rpc-laddr [value]",
	Short: "Set the rpc.laddr value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCLaddr,
}

func runSetRPCLaddr(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.Laddr = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.laddr set to: %s", newValue))
	return nil
}

// setRPCGRPCLaddrCmd sets the rpc.grpc_laddr in config.toml
var setRPCGRPCLaddrCmd = &cobra.Command{
	Use:   "rpc-grpc-laddr [value]",
	Short: "Set the rpc.grpc_laddr value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCGRPCLaddr,
}

func runSetRPCGRPCLaddr(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	// Load the current configuration
	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update the nested [rpc] section
	config.RPC.GRPCLaddr = newValue

	// Save the updated configuration
	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.grpc_laddr set to: %s", newValue))
	return nil
}

// setRPCMaxOpenConnectionsCmd sets the rpc.grpc_max_open_connections in config.toml
var setRPCMaxOpenConnectionsCmd = &cobra.Command{
	Use:   "rpc-grpc-max-open-connections [value]",
	Short: "Set the rpc.grpc_max_open_connections value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCMaxOpenConnections,
}

func runSetRPCMaxOpenConnections(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.grpc_max_open_connections value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.RPC.GRPCMaxOpenConnections = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.grpc_max_open_connections set to: %d", newValue))
	return nil
}

// setRPCUnsafeCmd sets the rpc.unsafe in config.toml
var setRPCUnsafeCmd = &cobra.Command{
	Use:   "rpc-unsafe [value]",
	Short: "Set the rpc.unsafe value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCUnsafe,
}

func runSetRPCUnsafe(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid rpc.unsafe value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.Unsafe = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.unsafe set to: %t", newValue))
	return nil
}

// setRPCMaxSubscriptionClientsCmd sets the rpc.max_subscription_clients in config.toml
var setRPCMaxSubscriptionClientsCmd = &cobra.Command{
	Use:   "rpc-max-subscription-clients [value]",
	Short: "Set the rpc.max_subscription_clients value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCMaxSubscriptionClients,
}

func runSetRPCMaxSubscriptionClients(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.max_subscription_clients value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.MaxSubscriptionClients = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.max_subscription_clients set to: %d", newValue))
	return nil
}

// setRPCMaxSubscriptionsPerClientCmd sets the rpc.max_subscriptions_per_client in config.toml
var setRPCMaxSubscriptionsPerClientCmd = &cobra.Command{
	Use:   "rpc-max-subscriptions-per-client [value]",
	Short: "Set the rpc.max_subscriptions_per_client value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCMaxSubscriptionsPerClient,
}

func runSetRPCMaxSubscriptionsPerClient(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.max_subscriptions_per_client value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.MaxSubscriptionsPerClient = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.max_subscriptions_per_client set to: %d", newValue))
	return nil
}

// setRPCExperimentalSubscriptionBufferSizeCmd sets the rpc.experimental_subscription_buffer_size in config.toml
var setRPCExperimentalSubscriptionBufferSizeCmd = &cobra.Command{
	Use:   "rpc-experimental-subscription-buffer-size [value]",
	Short: "Set the rpc.experimental_subscription_buffer_size value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCExperimentalSubscriptionBufferSize,
}

func runSetRPCExperimentalSubscriptionBufferSize(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.experimental_subscription_buffer_size value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.ExperimentalSubscriptionBufferSize = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.experimental_subscription_buffer_size set to: %d", newValue))
	return nil
}

// setRPCExperimentalWebsocketWriteBufferSizeCmd sets the rpc.experimental_websocket_write_buffer_size in config.toml
var setRPCExperimentalWebsocketWriteBufferSizeCmd = &cobra.Command{
	Use:   "rpc-experimental-websocket-write-buffer-size [value]",
	Short: "Set the rpc.experimental_websocket_write_buffer_size value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCExperimentalWebsocketWriteBufferSize,
}

func runSetRPCExperimentalWebsocketWriteBufferSize(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.experimental_websocket_write_buffer_size value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.ExperimentalWebsocketWriteBufferSize = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.experimental_websocket_write_buffer_size set to: %d", newValue))
	return nil
}

// setRPCExperimentalCloseOnSlowClientCmd sets the rpc.experimental_close_on_slow_client in config.toml
var setRPCExperimentalCloseOnSlowClientCmd = &cobra.Command{
	Use:   "rpc-experimental-close-on-slow-client [value]",
	Short: "Set the rpc.experimental_close_on_slow_client value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCExperimentalCloseOnSlowClient,
}

func runSetRPCExperimentalCloseOnSlowClient(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid rpc.experimental_close_on_slow_client value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.ExperimentalCloseOnSlowClient = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.experimental_close_on_slow_client set to: %t", newValue))
	return nil
}

// setRPCTimeoutBroadcastTxCommitCmd sets the rpc.timeout_broadcast_tx_commit in config.toml
var setRPCTimeoutBroadcastTxCommitCmd = &cobra.Command{
	Use:   "rpc-timeout-broadcast-tx-commit [value]",
	Short: "Set the rpc.timeout_broadcast_tx_commit value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCTimeoutBroadcastTxCommit,
}

func runSetRPCTimeoutBroadcastTxCommit(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.TimeoutBroadcastTxCommit = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.timeout_broadcast_tx_commit set to: %s", newValue))
	return nil
}

// setRPCMaxRequestBatchSizeCmd sets the rpc.max_request_batch_size in config.toml
var setRPCMaxRequestBatchSizeCmd = &cobra.Command{
	Use:   "rpc-max-request-batch-size [value]",
	Short: "Set the rpc.max_request_batch_size value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCMaxRequestBatchSize,
}

func runSetRPCMaxRequestBatchSize(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.max_request_batch_size value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.MaxRequestBatchSize = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.max_request_batch_size set to: %d", newValue))
	return nil
}

// setRPCMaxBodyBytesCmd sets the rpc.max_body_bytes in config.toml
var setRPCMaxBodyBytesCmd = &cobra.Command{
	Use:   "rpc-max-body-bytes [value]",
	Short: "Set the rpc.max_body_bytes value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCMaxBodyBytes,
}

func runSetRPCMaxBodyBytes(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.max_body_bytes value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.MaxBodyBytes = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.max_body_bytes set to: %d", newValue))
	return nil
}

// setRPCMaxHeaderBytesCmd sets the rpc.max_header_bytes in config.toml
var setRPCMaxHeaderBytesCmd = &cobra.Command{
	Use:   "rpc-max-header-bytes [value]",
	Short: "Set the rpc.max_header_bytes value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCMaxHeaderBytes,
}

func runSetRPCMaxHeaderBytes(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid rpc.max_header_bytes value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.MaxHeaderBytes = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.max_header_bytes set to: %d", newValue))
	return nil
}

// setRPCPPROFLaddrCmd sets the rpc.pprof_laddr in config.toml
var setRPCPPROFLaddrCmd = &cobra.Command{
	Use:   "rpc-pprof-laddr [value]",
	Short: "Set the rpc.pprof_laddr value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetRPCPPROFLaddr,
}

func runSetRPCPPROFLaddr(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [rpc] section
	config.RPC.PPROFLaddr = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("rpc.pprof_laddr set to: %s", newValue))
	return nil
}

// ----------------------- //
// [p2p] section set commands
// ----------------------- //

// setP2PLaddrCmd sets the p2p.laddr in config.toml
var setP2PLaddrCmd = &cobra.Command{
	Use:   "p2p-laddr [value]",
	Short: "Set the p2p.laddr value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PLaddr,
}

func runSetP2PLaddr(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.Laddr = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.laddr set to: %s", newValue))
	return nil
}

// setP2PExternalAddressCmd sets the p2p.external_address in config.toml
var setP2PExternalAddressCmd = &cobra.Command{
	Use:   "p2p-external-address [value]",
	Short: "Set the p2p.external_address value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PExternalAddress,
}

func runSetP2PExternalAddress(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.ExternalAddress = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.external_address set to: %s", newValue))
	return nil
}

// setP2PSeedsCmd sets the p2p.seeds in config.toml
var setP2PSeedsCmd = &cobra.Command{
	Use:   "p2p-seeds [value]",
	Short: "Set the p2p.seeds value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PSeeds,
}

func runSetP2PSeeds(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.Seeds = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.seeds set to: %s", newValue))
	return nil
}

// setP2PPersistentPeersCmd sets the p2p.persistent_peers in config.toml
var setP2PPersistentPeersCmd = &cobra.Command{
	Use:   "p2p-persistent-peers [value]",
	Short: "Set the p2p.persistent_peers value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PPersistentPeers,
}

func runSetP2PPersistentPeers(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.PersistentPeers = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.persistent_peers set to: %s", newValue))
	return nil
}

// setP2PAddrBookFileCmd sets the p2p.addr_book_file in config.toml
var setP2PAddrBookFileCmd = &cobra.Command{
	Use:   "p2p-addr-book-file [value]",
	Short: "Set the p2p.addr_book_file value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PAddrBookFile,
}

func runSetP2PAddrBookFile(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.AddrBookFile = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.addr_book_file set to: %s", newValue))
	return nil
}

// setP2PAddrBookStrictCmd sets the p2p.addr_book_strict in config
// setP2PAddrBookStrictCmd sets the p2p.addr_book_strict in config.toml
var setP2PAddrBookStrictCmd = &cobra.Command{
	Use:   "p2p-addr-book-strict [value]",
	Short: "Set the p2p.addr_book_strict value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PAddrBookStrict,
}

func runSetP2PAddrBookStrict(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid p2p.addr_book_strict value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.AddrBookStrict = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.addr_book_strict set to: %t", newValue))
	return nil
}

// setP2PMaxNumInboundPeersCmd sets the p2p.max_num_inbound_peers in config.toml
var setP2PMaxNumInboundPeersCmd = &cobra.Command{
	Use:   "p2p-max-num-inbound-peers [value]",
	Short: "Set the p2p.max_num_inbound_peers value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PMaxNumInboundPeers,
}

func runSetP2PMaxNumInboundPeers(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid p2p.max_num_inbound_peers value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.MaxNumInboundPeers = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.max_num_inbound_peers set to: %d", newValue))
	return nil
}

// setP2PMaxNumOutboundPeersCmd sets the p2p.max_num_outbound_peers in config.toml
var setP2PMaxNumOutboundPeersCmd = &cobra.Command{
	Use:   "p2p-max-num-outbound-peers [value]",
	Short: "Set the p2p.max_num_outbound_peers value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PMaxNumOutboundPeers,
}

func runSetP2PMaxNumOutboundPeers(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid p2p.max_num_outbound_peers value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.MaxNumOutboundPeers = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.max_num_outbound_peers set to: %d", newValue))
	return nil
}

// setP2PUnconditionalPeerIDsCmd sets the p2p.unconditional_peer_ids in config.toml
var setP2PUnconditionalPeerIDsCmd = &cobra.Command{
	Use:   "p2p-unconditional-peer-ids [value]",
	Short: "Set the p2p.unconditional_peer_ids value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PUnconditionalPeerIDs,
}

func runSetP2PUnconditionalPeerIDs(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.UnconditionalPeerIDs = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.unconditional_peer_ids set to: %s", newValue))
	return nil
}

// setP2PPersistentPeersMaxDialPeriodCmd sets the p2p.persistent_peers_max_dial_period in config.toml
var setP2PPersistentPeersMaxDialPeriodCmd = &cobra.Command{
	Use:   "p2p-persistent-peers-max-dial-period [value]",
	Short: "Set the p2p.persistent_peers_max_dial_period value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PPersistentPeersMaxDialPeriod,
}

func runSetP2PPersistentPeersMaxDialPeriod(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.PersistentPeersMaxDialPeriod = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.persistent_peers_max_dial_period set to: %s", newValue))
	return nil
}

// setP2PFlushThrottleTimeoutCmd sets the p2p.flush_throttle_timeout in config.toml
var setP2PFlushThrottleTimeoutCmd = &cobra.Command{
	Use:   "p2p-flush-throttle-timeout [value]",
	Short: "Set the p2p.flush_throttle_timeout value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PFlushThrottleTimeout,
}

func runSetP2PFlushThrottleTimeout(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.FlushThrottleTimeout = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.flush_throttle_timeout set to: %s", newValue))
	return nil
}

// setP2PMaxPacketMsgPayloadSizeCmd sets the p2p.max_packet_msg_payload_size in config.toml
var setP2PMaxPacketMsgPayloadSizeCmd = &cobra.Command{
	Use:   "p2p-max-packet-msg-payload-size [value]",
	Short: "Set the p2p.max_packet_msg_payload_size value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PMaxPacketMsgPayloadSize,
}

func runSetP2PMaxPacketMsgPayloadSize(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid p2p.max_packet_msg_payload_size value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.MaxPacketMsgPayloadSize = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.max_packet_msg_payload_size set to: %d", newValue))
	return nil
}

// setP2PSendRateCmd sets the p2p.send_rate in config.toml
var setP2PSendRateCmd = &cobra.Command{
	Use:   "p2p-send-rate [value]",
	Short: "Set the p2p.send_rate value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PSendRate,
}

func runSetP2PSendRate(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid p2p.send_rate value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.SendRate = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.send_rate set to: %d", newValue))
	return nil
}

// setP2PRecvRateCmd sets the p2p.recv_rate in config.toml
var setP2PRecvRateCmd = &cobra.Command{
	Use:   "p2p-recv-rate [value]",
	Short: "Set the p2p.recv_rate value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PRecvRate,
}

func runSetP2PRecvRate(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid p2p.recv_rate value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.RecvRate = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.recv_rate set to: %d", newValue))
	return nil
}

// setP2PPEXCmd sets the p2p.pex in config.toml
var setP2PPEXCmd = &cobra.Command{
	Use:   "p2p-pex [value]",
	Short: "Set the p2p.pex value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PPEX,
}

func runSetP2PPEX(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid p2p.pex value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.PEX = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.pex set to: %t", newValue))
	return nil
}

// setP2PSeedModeCmd sets the p2p.seed_mode in config.toml
var setP2PSeedModeCmd = &cobra.Command{
	Use:   "p2p-seed-mode [value]",
	Short: "Set the p2p.seed_mode value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PSeedMode,
}

func runSetP2PSeedMode(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid p2p.seed_mode value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.SeedMode = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.seed_mode set to: %t", newValue))
	return nil
}

// setP2PAllowDuplicateIPCmd sets the p2p.allow_duplicate_ip in config.toml
var setP2PAllowDuplicateIPCmd = &cobra.Command{
	Use:   "p2p-allow-duplicate-ip [value]",
	Short: "Set the p2p.allow_duplicate_ip value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PAllowDuplicateIP,
}

func runSetP2PAllowDuplicateIP(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid p2p.allow_duplicate_ip value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.AllowDuplicateIP = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.allow_duplicate_ip set to: %t", newValue))
	return nil
}

// setP2PHandshakeTimeoutCmd sets the p2p.handshake_timeout in config.toml
var setP2PHandshakeTimeoutCmd = &cobra.Command{
	Use:   "p2p-handshake-timeout [value]",
	Short: "Set the p2p.handshake_timeout value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PHandshakeTimeout,
}

func runSetP2PHandshakeTimeout(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.HandshakeTimeout = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.handshake_timeout set to: %s", newValue))
	return nil
}

// setP2PDialTimeoutCmd sets the p2p.dial_timeout in config.toml
var setP2PDialTimeoutCmd = &cobra.Command{
	Use:   "p2p-dial-timeout [value]",
	Short: "Set the p2p.dial_timeout value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetP2PDialTimeout,
}

func runSetP2PDialTimeout(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	// Update nested [p2p] section
	config.P2P.DialTimeout = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("p2p.dial_timeout set to: %s", newValue))
	return nil
}

// ----------------------- //
// [mempool] section set commands
// ----------------------- //

// setMempoolTypeCmd sets the mempool.type in config.toml
var setMempoolTypeCmd = &cobra.Command{
	Use:   "mempool-type [value]",
	Short: "Set the mempool.type value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolType,
}

func runSetMempoolType(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.Type = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.type set to: %s", newValue))
	return nil
}

// setMempoolRecheckCmd sets the mempool.recheck in config.toml
var setMempoolRecheckCmd = &cobra.Command{
	Use:   "mempool-recheck [value]",
	Short: "Set the mempool.recheck value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolRecheck,
}

func runSetMempoolRecheck(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid mempool.recheck value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.Recheck = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.recheck set to: %t", newValue))
	return nil
}

// setMempoolRecheckTimeoutCmd sets the mempool.recheck_timeout in config.toml
var setMempoolRecheckTimeoutCmd = &cobra.Command{
	Use:   "mempool-recheck-timeout [value]",
	Short: "Set the mempool.recheck_timeout value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolRecheckTimeout,
}

func runSetMempoolRecheckTimeout(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.RecheckTimeout = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.recheck_timeout set to: %s", newValue))
	return nil
}

// setMempoolBroadcastCmd sets the mempool.broadcast in config.toml
var setMempoolBroadcastCmd = &cobra.Command{
	Use:   "mempool-broadcast [value]",
	Short: "Set the mempool.broadcast value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolBroadcast,
}

func runSetMempoolBroadcast(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid mempool.broadcast value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.Broadcast = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.broadcast set to: %t", newValue))
	return nil
}

// setMempoolSizeCmd sets the mempool.size in config.toml
var setMempoolSizeCmd = &cobra.Command{
	Use:   "mempool-size [value]",
	Short: "Set the mempool.size value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolSize,
}

func runSetMempoolSize(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid mempool.size value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.Size = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.size set to: %d", newValue))
	return nil
}

// setMempoolMaxTxsBytesCmd sets the mempool.max_txs_bytes in config.toml
var setMempoolMaxTxsBytesCmd = &cobra.Command{
	Use:   "mempool-max-txs-bytes [value]",
	Short: "Set the mempool.max_txs_bytes value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolMaxTxsBytes,
}

func runSetMempoolMaxTxsBytes(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid mempool.max_txs_bytes value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.MaxTxsBytes = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.max_txs_bytes set to: %d", newValue))
	return nil
}

// setMempoolCacheSizeCmd sets the mempool.cache_size in config.toml
var setMempoolCacheSizeCmd = &cobra.Command{
	Use:   "mempool-cache-size [value]",
	Short: "Set the mempool.cache_size value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolCacheSize,
}

func runSetMempoolCacheSize(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid mempool.cache_size value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.CacheSize = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.cache_size set to: %d", newValue))
	return nil
}

// setMempoolKeepInvalidTxsInCacheCmd sets the mempool.keep_invalid_txs_in_cache in config.toml
var setMempoolKeepInvalidTxsInCacheCmd = &cobra.Command{
	Use:   "mempool-keep-invalid-txs-in-cache [value]",
	Short: "Set the mempool.keep_invalid_txs_in_cache value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolKeepInvalidTxsInCache,
}

func runSetMempoolKeepInvalidTxsInCache(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid mempool.keep_invalid_txs_in_cache value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.KeepInvalidTxsInCache = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.keep_invalid_txs_in_cache set to: %t", newValue))
	return nil
}

// setMempoolMaxTxBytesCmd sets the mempool.max_tx_bytes in config.toml
var setMempoolMaxTxBytesCmd = &cobra.Command{
	Use:   "mempool-max-tx-bytes [value]",
	Short: "Set the mempool.max_tx_bytes value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolMaxTxBytes,
}

func runSetMempoolMaxTxBytes(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid mempool.max_tx_bytes value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.MaxTxBytes = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.max_tx_bytes set to: %d", newValue))
	return nil
}

// setMempoolMaxBatchBytesCmd sets the mempool.max_batch_bytes in config.toml
var setMempoolMaxBatchBytesCmd = &cobra.Command{
	Use:   "mempool-max-batch-bytes [value]",
	Short: "Set the mempool.max_batch_bytes value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetMempoolMaxBatchBytes,
}

func runSetMempoolMaxBatchBytes(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid mempool.max_batch_bytes value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Mempool.MaxBatchBytes = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("mempool.max_batch_bytes set to: %d", newValue))
	return nil
}

// ----------------------- //
// [statesync] section set commands
// ----------------------- //

// setStatesyncEnableCmd sets the statesync.enable in config.toml
var setStatesyncEnableCmd = &cobra.Command{
	Use:   "statesync-enable [value]",
	Short: "Set the statesync.enable value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncEnable,
}

func runSetStatesyncEnable(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid statesync.enable value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.Enable = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.enable set to: %t", newValue))
	return nil
}

// setStatesyncRPCServersCmd sets the statesync.rpc_servers in config.toml
var setStatesyncRPCServersCmd = &cobra.Command{
	Use:   "statesync-rpc-servers [value]",
	Short: "Set the statesync.rpc_servers value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncRPCServers,
}

func runSetStatesyncRPCServers(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.RPCServers = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.rpc_servers set to: %s", newValue))
	return nil
}

// setStatesyncTrustHeightCmd sets the statesync.trust_height in config.toml
var setStatesyncTrustHeightCmd = &cobra.Command{
	Use:   "statesync-trust-height [value]",
	Short: "Set the statesync.trust_height value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncTrustHeight,
}

func runSetStatesyncTrustHeight(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid statesync.trust_height value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.TrustHeight = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.trust_height set to: %d", newValue))
	return nil
}

// setStatesyncTrustHashCmd sets the statesync.trust_hash in config.toml
var setStatesyncTrustHashCmd = &cobra.Command{
	Use:   "statesync-trust-hash [value]",
	Short: "Set the statesync.trust_hash value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncTrustHash,
}

func runSetStatesyncTrustHash(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.TrustHash = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.trust_hash set to: %s", newValue))
	return nil
}

// setStatesyncTrustPeriodCmd sets the statesync.trust_period in config.toml
var setStatesyncTrustPeriodCmd = &cobra.Command{
	Use:   "statesync-trust-period [value]",
	Short: "Set the statesync.trust_period value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncTrustPeriod,
}

func runSetStatesyncTrustPeriod(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.TrustPeriod = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.trust_period set to: %s", newValue))
	return nil
}

// setStatesyncDiscoveryTimeCmd sets the statesync.discovery_time in config.toml
var setStatesyncDiscoveryTimeCmd = &cobra.Command{
	Use:   "statesync-discovery-time [value]",
	Short: "Set the statesync.discovery_time value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncDiscoveryTime,
}

func runSetStatesyncDiscoveryTime(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.DiscoveryTime = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.discovery_time set to: %s", newValue))
	return nil
}

// setStatesyncTempDirCmd sets the statesync.temp_dir in config.toml
var setStatesyncTempDirCmd = &cobra.Command{
	Use:   "statesync-temp-dir [value]",
	Short: "Set the statesync.temp_dir value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncTempDir,
}

func runSetStatesyncTempDir(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.TempDir = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.temp_dir set to: %s", newValue))
	return nil
}

// setStatesyncChunkRequestTimeoutCmd sets the statesync.chunk_request_timeout in config.toml
var setStatesyncChunkRequestTimeoutCmd = &cobra.Command{
	Use:   "statesync-chunk-request-timeout [value]",
	Short: "Set the statesync.chunk_request_timeout value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncChunkRequestTimeout,
}

func runSetStatesyncChunkRequestTimeout(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.ChunkRequestTimeout = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.chunk_request_timeout set to: %s", newValue))
	return nil
}

// setStatesyncChunkFetchersCmd sets the statesync.chunk_fetchers in config.toml
var setStatesyncChunkFetchersCmd = &cobra.Command{
	Use:   "statesync-chunk-fetchers [value]",
	Short: "Set the statesync.chunk_fetchers value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStatesyncChunkFetchers,
}

func runSetStatesyncChunkFetchers(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid statesync.chunk_fetchers value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Statesync.ChunkFetchers = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("statesync.chunk_fetchers set to: %d", newValue))
	return nil
}

// ----------------------- //
// [blocksync] section set commands
// ----------------------- //

// setBlocksyncVersionCmd sets the blocksync.version in config.toml
var setBlocksyncVersionCmd = &cobra.Command{
	Use:   "blocksync-version [value]",
	Short: "Set the blocksync.version value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetBlocksyncVersion,
}

func runSetBlocksyncVersion(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Blocksync.Version = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("blocksync.version set to: %s", newValue))
	return nil
}

// ----------------------- //
// [consensus] section set commands
// ----------------------- //

// setConsensusWALFileCmd sets the consensus.wal_file in config.toml
var setConsensusWALFileCmd = &cobra.Command{
	Use:   "consensus-wal-file [value]",
	Short: "Set the consensus.wal_file value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusWALFile,
}

func runSetConsensusWALFile(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.WALFile = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.wal_file set to: %s", newValue))
	return nil
}

// setConsensusTimeoutProposeCmd sets the consensus.timeout_propose in config.toml
var setConsensusTimeoutProposeCmd = &cobra.Command{
	Use:   "consensus-timeout-propose [value]",
	Short: "Set the consensus.timeout_propose value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusTimeoutPropose,
}

func runSetConsensusTimeoutPropose(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.TimeoutPropose = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.timeout_propose set to: %s", newValue))
	return nil
}

// setConsensusTimeoutProposeDeltaCmd sets the consensus.timeout_propose_delta in config.toml
var setConsensusTimeoutProposeDeltaCmd = &cobra.Command{
	Use:   "consensus-timeout-propose-delta [value]",
	Short: "Set the consensus.timeout_propose_delta value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusTimeoutProposeDelta,
}

func runSetConsensusTimeoutProposeDelta(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.TimeoutProposeDelta = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.timeout_propose_delta set to: %s", newValue))
	return nil
}

// setConsensusTimeoutPrevoteCmd sets the consensus.timeout_prevote in config.toml
var setConsensusTimeoutPrevoteCmd = &cobra.Command{
	Use:   "consensus-timeout-prevote [value]",
	Short: "Set the consensus.timeout_prevote value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusTimeoutPrevote,
}

func runSetConsensusTimeoutPrevote(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.TimeoutPrevote = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.timeout_prevote set to: %s", newValue))
	return nil
}

// setConsensusTimeoutPrevoteDeltaCmd sets the consensus.timeout_prevote_delta in config.toml
var setConsensusTimeoutPrevoteDeltaCmd = &cobra.Command{
	Use:   "consensus-timeout-prevote-delta [value]",
	Short: "Set the consensus.timeout_prevote_delta value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusTimeoutPrevoteDelta,
}

func runSetConsensusTimeoutPrevoteDelta(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.TimeoutPrevoteDelta = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.timeout_prevote_delta set to: %s", newValue))
	return nil
}

// setConsensusTimeoutPrecommitCmd sets the consensus.timeout_precommit in config.toml
var setConsensusTimeoutPrecommitCmd = &cobra.Command{
	Use:   "consensus-timeout-precommit [value]",
	Short: "Set the consensus.timeout_precommit value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusTimeoutPrecommit,
}

func runSetConsensusTimeoutPrecommit(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.TimeoutPrecommit = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.timeout_precommit set to: %s", newValue))
	return nil
}

// setConsensusTimeoutPrecommitDeltaCmd sets the consensus.timeout_precommit_delta in config.toml
var setConsensusTimeoutPrecommitDeltaCmd = &cobra.Command{
	Use:   "consensus-timeout-precommit-delta [value]",
	Short: "Set the consensus.timeout_precommit_delta value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusTimeoutPrecommitDelta,
}

func runSetConsensusTimeoutPrecommitDelta(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.TimeoutPrecommitDelta = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.timeout_precommit_delta set to: %s", newValue))
	return nil
}

// setConsensusTimeoutCommitCmd sets the consensus.timeout_commit in config.toml
var setConsensusTimeoutCommitCmd = &cobra.Command{
	Use:   "consensus-timeout-commit [value]",
	Short: "Set the consensus.timeout_commit value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusTimeoutCommit,
}

func runSetConsensusTimeoutCommit(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.TimeoutCommit = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.timeout_commit set to: %s", newValue))
	return nil
}

// setConsensusDoubleSignCheckHeightCmd sets the consensus.double_sign_check_height in config.toml
var setConsensusDoubleSignCheckHeightCmd = &cobra.Command{
	Use:   "consensus-double-sign-check-height [value]",
	Short: "Set the consensus.double_sign_check_height value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusDoubleSignCheckHeight,
}

func runSetConsensusDoubleSignCheckHeight(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid consensus.double_sign_check_height value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.DoubleSignCheckHeight = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.double_sign_check_height set to: %d", newValue))
	return nil
}

// setConsensusSkipTimeoutCommitCmd sets the consensus.skip_timeout_commit in config.toml
var setConsensusSkipTimeoutCommitCmd = &cobra.Command{
	Use:   "consensus-skip-timeout-commit [value]",
	Short: "Set the consensus.skip_timeout_commit value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusSkipTimeoutCommit,
}

func runSetConsensusSkipTimeoutCommit(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid consensus.skip_timeout_commit value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.SkipTimeoutCommit = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.skip_timeout_commit set to: %t", newValue))
	return nil
}

// setConsensusCreateEmptyBlocksCmd sets the consensus.create_empty_blocks in config.toml
var setConsensusCreateEmptyBlocksCmd = &cobra.Command{
	Use:   "consensus-create-empty-blocks [value]",
	Short: "Set the consensus.create_empty_blocks value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusCreateEmptyBlocks,
}

func runSetConsensusCreateEmptyBlocks(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid consensus.create_empty_blocks value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.CreateEmptyBlocks = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.create_empty_blocks set to: %t", newValue))
	return nil
}

// setConsensusCreateEmptyBlocksIntervalCmd sets the consensus.create_empty_blocks_interval in config.toml
var setConsensusCreateEmptyBlocksIntervalCmd = &cobra.Command{
	Use:   "consensus-create-empty-blocks-interval [value]",
	Short: "Set the consensus.create_empty_blocks_interval value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusCreateEmptyBlocksInterval,
}

func runSetConsensusCreateEmptyBlocksInterval(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.CreateEmptyBlocksInterval = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.create_empty_blocks_interval set to: %s", newValue))
	return nil
}

// setConsensusPeerGossipSleepDurationCmd sets the consensus.peer_gossip_sleep_duration in config.toml
var setConsensusPeerGossipSleepDurationCmd = &cobra.Command{
	Use:   "consensus-peer-gossip-sleep-duration [value]",
	Short: "Set the consensus.peer_gossip_sleep_duration value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusPeerGossipSleepDuration,
}

func runSetConsensusPeerGossipSleepDuration(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.PeerGossipSleepDuration = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.peer_gossip_sleep_duration set to: %s", newValue))
	return nil
}

// setConsensusPeerQueryMaj23SleepDurationCmd sets the consensus.peer_query_maj23_sleep_duration in config.toml
var setConsensusPeerQueryMaj23SleepDurationCmd = &cobra.Command{
	Use:   "consensus-peer-query-maj23-sleep-duration [value]",
	Short: "Set the consensus.peer_query_maj23_sleep_duration value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetConsensusPeerQueryMaj23SleepDuration,
}

func runSetConsensusPeerQueryMaj23SleepDuration(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Consensus.PeerQueryMaj23SleepDuration = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("consensus.peer_query_maj23_sleep_duration set to: %s", newValue))
	return nil
}

// ----------------------- //
// [storage] section set commands
// ----------------------- //

// setStorageDiscardABCIResponsesCmd sets the storage.discard_abci_responses in config.toml
var setStorageDiscardABCIResponsesCmd = &cobra.Command{
	Use:   "storage-discard-abci-responses [value]",
	Short: "Set the storage.discard_abci_responses value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetStorageDiscardABCIResponses,
}

func runSetStorageDiscardABCIResponses(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid storage.discard_abci_responses value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Storage.DiscardABCIResponses = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("storage.discard_abci_responses set to: %t", newValue))
	return nil
}

// ----------------------- //
// [tx_index] section set commands
// ----------------------- //

// setTxIndexIndexerCmd sets the tx_index.indexer in config.toml
var setTxIndexIndexerCmd = &cobra.Command{
	Use:   "tx-index-indexer [value]",
	Short: "Set the tx_index.indexer value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetTxIndexIndexer,
}

func runSetTxIndexIndexer(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.TxIndex.Indexer = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("tx_index.indexer set to: %s", newValue))
	return nil
}

var setTxIndexPSQLConnCmd = &cobra.Command{
	Use:   "tx-index-psql-conn [value]",
	Short: "Set the tx_index.psql-conn value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetTxIndexPSQLConn,
}

func runSetTxIndexPSQLConn(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.TxIndex.PSQLConn = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("tx_index.psql-conn set to: %s", newValue))
	return nil
}

// ----------------------- //
// [instrumentation] section set commands
// ----------------------- //

var setInstrumentationPrometheusCmd = &cobra.Command{
	Use:   "instrumentation-prometheus [value]",
	Short: "Set the instrumentation.prometheus value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetInstrumentationPrometheus,
}

func runSetInstrumentationPrometheus(cmd *cobra.Command, args []string) error {
	input := args[0]
	newValue, err := strconv.ParseBool(input)
	if err != nil {
		return fmt.Errorf("invalid instrumentation.prometheus value: %s. Must be true or false", input)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Instrumentation.Prometheus = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("instrumentation.prometheus set to: %t", newValue))
	return nil
}

// setInstrumentationPrometheusListenAddrCmd sets the instrumentation.prometheus_listen_addr in config.toml
var setInstrumentationPrometheusListenAddrCmd = &cobra.Command{
	Use:   "instrumentation-prometheus-listen-addr [value]",
	Short: "Set the instrumentation.prometheus_listen_addr value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetInstrumentationPrometheusListenAddr,
}

func runSetInstrumentationPrometheusListenAddr(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Instrumentation.PrometheusListenAddr = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("instrumentation.prometheus_listen_addr set to: %s", newValue))
	return nil
}

// setInstrumentationMaxOpenConnectionsCmd sets the instrumentation.max_open_connections in config.toml
var setInstrumentationMaxOpenConnectionsCmd = &cobra.Command{
	Use:   "instrumentation-max-open-connections [value]",
	Short: "Set the instrumentation.max_open_connections value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetInstrumentationMaxOpenConnections,
}

func runSetInstrumentationMaxOpenConnections(cmd *cobra.Command, args []string) error {
	newValueStr := args[0]
	newValue, err := strconv.Atoi(newValueStr)
	if err != nil {
		return fmt.Errorf("invalid instrumentation.max_open_connections value: %s. Must be an integer", newValueStr)
	}

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Instrumentation.MaxOpenConnections = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("instrumentation.max_open_connections set to: %d", newValue))
	return nil
}

// setInstrumentationNamespaceCmd sets the instrumentation.namespace in config.toml
var setInstrumentationNamespaceCmd = &cobra.Command{
	Use:   "instrumentation-namespace [value]",
	Short: "Set the instrumentation.namespace value in config.toml",
	Args:  cobra.ExactArgs(1),
	RunE:  runSetInstrumentationNamespace,
}

func runSetInstrumentationNamespace(cmd *cobra.Command, args []string) error {
	newValue := args[0]

	config, err := loadSetConfigConfig()
	if err != nil {
		return err
	}

	config.Instrumentation.Namespace = newValue

	if err := saveSetConfigConfig(config); err != nil {
		return err
	}

	printInfo(fmt.Sprintf("instrumentation.namespace set to: %s", newValue))
	return nil
}

// loadSetStoryConfig loads the story.toml configuration file
func loadSetStoryConfig() (*SetStoryConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get user home directory: %v", err))
		return nil, err
	}

	storyTomlPath := filepath.Join(homeDir, ".story", "story", "config", "story.toml")

	data, err := os.ReadFile(storyTomlPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to load story.toml: %v", err))
		return nil, err
	}

	var config SetStoryConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		printError(fmt.Sprintf("Failed to parse story.toml: %v", err))
		return nil, err
	}

	return &config, nil
}

// loadSetConfigConfig loads the config.toml configuration file into the SetConfigConfig struct
func loadSetConfigConfig() (*SetConfigConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get user home directory: %v", err))
		return nil, err
	}

	configTomlPath := filepath.Join(homeDir, ".story", "story", "config", "config.toml")

	data, err := os.ReadFile(configTomlPath)
	if err != nil {
		printError(fmt.Sprintf("Failed to load config.toml: %v", err))
		return nil, err
	}

	var config SetConfigConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		printError(fmt.Sprintf("Failed to parse config.toml: %v", err))
		return nil, err
	}

	return &config, nil
}

// saveSetStoryConfig saves the updated story.toml configuration file
func saveSetStoryConfig(config *SetStoryConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get user home directory: %v", err))
		return err
	}

	storyTomlPath := filepath.Join(homeDir, ".story", "story", "config", "story.toml")

	// Backup the original file
	if err := backupFile(storyTomlPath); err != nil {
		return fmt.Errorf("failed to backup story.toml: %v", err)
	}

	// Marshal the struct back to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		printError(fmt.Sprintf("Failed to marshal story.toml: %v", err))
		return err
	}

	// Write back to the file
	if err := os.WriteFile(storyTomlPath, data, 0644); err != nil {
		printError(fmt.Sprintf("Failed to write to story.toml: %v", err))
		return err
	}

	return nil
}

// saveSetConfigConfig saves the updated SetConfigConfig struct back to config.toml
func saveSetConfigConfig(config *SetConfigConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		printError(fmt.Sprintf("Failed to get user home directory: %v", err))
		return err
	}

	configTomlPath := filepath.Join(homeDir, ".story", "story", "config", "config.toml")

	// Backup the original file
	if err := backupFile(configTomlPath); err != nil {
		return fmt.Errorf("failed to backup config.toml: %v", err)
	}

	// Marshal the struct back to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		printError(fmt.Sprintf("Failed to marshal config.toml: %v", err))
		return err
	}

	// Write back to the file
	if err := os.WriteFile(configTomlPath, data, 0644); err != nil {
		printError(fmt.Sprintf("Failed to write to config.toml: %v", err))
		return err
	}

	return nil
}

// backupFile creates a timestamped backup of the given file
func backupFile(filePath string) error {
	timestamp := time.Now().Format("20060102_150405")
	backupPath := fmt.Sprintf("%s.bak.%s", filePath, timestamp)

	srcFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	printInfo(fmt.Sprintf("Backup created at: %s", backupPath))
	return nil
}

// printError prints error messages in red
func printError(message string) {
	c := color.New(color.FgRed).SprintFunc()
	fmt.Printf("ERROR: %s\n", c(message))
}

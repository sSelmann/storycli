package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"

	"github.com/sSelmann/storycli/cmd/snapshot"
	"github.com/sSelmann/storycli/utils/bash"
	"github.com/sSelmann/storycli/utils/config"
)

var (
	moniker    string
	customPort string
)

var (
	prunedSnapshotSize  string
	archiveSnapshotSize string
	pruningMode         string
	setupType           string
)

var recommendedCPU = 4
var recommendedRAM = 16 * 1024                 // in MB
var recommendedDisk = 200 * 1024 * 1024 * 1024 // in bytes

// setupNodeCmd represents the setup node command
var setupNodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Set up a new Story node",
	RunE:  runSetupNode,
}

func init() {
	setupNodeCmd.Flags().StringVar(&moniker, "moniker", "", "Your node's moniker")
	setupNodeCmd.Flags().StringVar(&customPort, "customport", "", "First two digits of the custom port (default: 26)")
	setupNodeCmd.Flags().StringVar(&pruningMode, "pruning-mode", "", "Pruning mode to use (pruned or archive)")
}

func runSetupNode(cmd *cobra.Command, args []string) error {
	// Step 0: System Resource Check
	err := checkSystemResources()
	if err != nil {
		return err
	}

	// Step 1: Check for existing Story installation
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	storyDir := fmt.Sprintf("%s/.story", homeDir)
	storyRepoDir := fmt.Sprintf("%s/story", homeDir)
	if _, err := os.Stat(storyDir); err == nil {
		// Directory exists
		prompt := promptui.Select{
			Label:     "It looks like you already have a Story node installed. Do you want to remove the existing installation?",
			Items:     []string{"Yes", "No"},
			CursorPos: 1,
		}
		_, result, err := prompt.Run()
		if err != nil {
			return err
		}

		if strings.ToLower(result) == "yes" {
			// Check if the services exist and are running
			pterm.Info.Println("Checking if Story and Story-Geth services are active...")
			storyActive := checkServiceExistsAndActive("story")
			storyGethActive := checkServiceExistsAndActive("story-geth")

			pterm.Info.Println("Stopping Story services...")
			if storyActive {
				err = bash.RunCommand("sudo", "systemctl", "stop", "story")
				if err != nil {
					return fmt.Errorf("failed to stop Story service: %v", err)
				}
			}

			if storyGethActive {
				err = bash.RunCommand("sudo", "systemctl", "stop", "story-geth")
				if err != nil {
					return fmt.Errorf("failed to stop Story-Geth service: %v", err)
				}
			}

			// Remove directories
			err := os.RemoveAll(storyDir)
			if err != nil {
				return fmt.Errorf("failed to remove Story directory: %v", err)
			}
			err = os.RemoveAll(storyRepoDir)
			if err != nil {
				return fmt.Errorf("failed to remove Story repository directory: %v", err)
			}
			pterm.Success.Println("Existing installation removed.")
		} else {
			pterm.Warning.Println("Setup aborted.")
			return nil
		}
	}

	// Step 2: Fetch Snapshot Sizes from API (only for Krews)
	//err = fetchSnapshotSizes()
	//if err != nil {
	//	pterm.Warning.Printf(fmt.Sprintf("Failed to fetch snapshot sizes: %v", err))
	//	// Continue without snapshot size info
	//}

	// Step 3: Prompt for missing flags
	if moniker == "" {
		prompt := promptui.Prompt{
			Label: "Enter your moniker",
		}
		moniker, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if customPort == "" {
		defaultPort := "26"
		prompt := promptui.Prompt{
			Label: "Enter the first two digits of the custom port (default 26): ",
			Validate: func(input string) error {
				if input == "" {
					return nil // Allow empty input to accept default
				}
				if len(input) > 2 {
					return errors.New("custom port must be a maximum of 2 digits")
				}
				if _, err := strconv.Atoi(input); err != nil {
					return errors.New("custom port must be numeric")
				}
				return nil
			},
		}
		customPort, err = prompt.Run()
		if err != nil {
			return err
		}
		if customPort == "" {
			customPort = defaultPort
		}
	}

	// Select setup type
	setupType, err := selectSetupType()
	if err != nil {
		pterm.Warning.Println(fmt.Sprintf("Failed to select setup type: %v", err))
	}

	// Pruning mode info message
	snapshot.PruningModeInformation()

	pruningMode, err := snapshot.SelectPruningMode()
	if err != nil {
		pterm.Warning.Println(fmt.Sprintf("Failed to select pruning mode: %v", err))
	}

	// Proceed with setup without Cosmovisor
	if setupType == "without cosmovisor" {
		err = setupWithoutCosmovisor(moniker, customPort, pruningMode)
		if err != nil {
			return err
		}
	} else if setupType == "with cosmovisor" {
		err = setupWithCosmovisor(moniker, customPort, pruningMode)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("failed to get correct setup type value: %s", err)
	}

	return nil
}

func checkServiceExistsAndActive(serviceName string) bool {
	cmd := exec.Command("systemctl", "is-active", "--quiet", serviceName)
	err := cmd.Run()
	return err == nil // Eğer `systemctl is-active` başarılı dönerse, servis aktif demektir
}

func checkSystemResources() error {
	pterm.Info.Println("Checking system resources...")

	// CPU cores
	cpuCores := runtime.NumCPU()
	if cpuCores < recommendedCPU {
		pterm.Warning.Println(fmt.Sprintf("You have %d CPU cores. Recommended is %d cores.", cpuCores, recommendedCPU))
	} else {
		pterm.Info.Println(fmt.Sprintf("CPU cores: %d", cpuCores))
	}

	// RAM
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	ramMB := vmStat.Total / (1024 * 1024)
	if int(ramMB) < recommendedRAM {
		pterm.Warning.Println(fmt.Sprintf("You have %d MB of RAM. Recommended is %d MB.", ramMB, recommendedRAM))
	} else {
		pterm.Info.Println(fmt.Sprintf("RAM: %d MB", ramMB))
	}

	// Disk space
	diskStat, err := disk.Usage("/")
	if err != nil {
		return err
	}
	diskGB := diskStat.Total / (1024 * 1024 * 1024)
	recommendedDiskGB := recommendedDisk / (1024 * 1024 * 1024)
	if diskStat.Total < uint64(recommendedDisk) {
		pterm.Warning.Println(fmt.Sprintf("You have %d GB of disk space. Recommended is %d GB.", diskGB, recommendedDiskGB))
	} else {
		pterm.Info.Println(fmt.Sprintf("Disk space: %d GB", diskGB))
	}

	return nil
}

func getLatestReleaseTag(repo string) (string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch latest release: %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", err
	}

	if release.TagName == "" {
		return "", errors.New("no tag_name found in latest release")
	}

	return release.TagName, nil
}

func setupWithoutCosmovisor(moniker, customPort, pruningMode string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	os.Setenv("MONIKER", moniker)
	os.Setenv("STORY_PORT", customPort)
	os.Setenv("PRUNING_MODE", pruningMode)

	// Navigate to home directory
	pterm.Info.Println("Navigating to home directory...")
	err = os.Chdir(homeDir)
	if err != nil {
		return err
	}

	// Download geth binary
	downloadBinary("geth", homeDir)

	// Create necessary directories
	createNecessaryDirectories()

	// Download story binary
	downloadBinary("story", homeDir)

	// Initialize Story
	initializingStoryNode(homeDir, moniker)

	// Configure seeds and peers
	err = configureSeedsAndPeers(homeDir)
	if err != nil {
		return err
	}

	// Download genesis and addrbook
	err = downloadGenesisAndAddrbook(homeDir)
	if err != nil {
		return err
	}

	// Set config.toml path
	configToml := fmt.Sprintf("%s/.story/story/config/config.toml", homeDir)

	// Set custom ports in story.toml
	settingCustomPorts(homeDir, configToml)

	// Enable Prometheus
	enablePrometheus(configToml)

	// Set indexer mode if pruned
	setIndexer(configToml, pruningMode)

	// Create systemd service files
	pterm.Info.Println("Creating systemd service files...")
	err = createServiceFilesWithoutCosmovisor(homeDir, customPort)
	if err != nil {
		return err
	}

	// Enable and start services
	pterm.Info.Println("Enabling services...")
	err = bash.RunCommand("sudo", "systemctl", "daemon-reload")
	if err != nil {
		return err
	}

	// Download snapshot based on provider
	pterm.Info.Println("Downloading snapshot...")
	snapshot.CallRunDownloadSnapshotManually(pruningMode, homeDir, false)

	pterm.Success.Println("Node setup without Cosmovisor completed successfully.")

	return nil
}

func configureSeedsAndPeers(homeDir string) error {
	pterm.Info.Println("Configuring seeds and peers...")

	seeds := "51ff395354c13fab493a03268249a74860b5f9cc@story-testnet-seed.itrocket.net:26656"

	// Fetch peers
	cmd := exec.Command("bash", "-c", `curl -sS https://story-testnet-rpc.itrocket.net/net_info | jq -r '.result.peers[] | "\(.node_info.id)@\(.remote_ip):\(.node_info.listen_addr)"' | awk -F ':' '{print $1":"$(NF)}' | paste -sd, -`)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	peers := strings.TrimSpace(string(out))

	configFile := fmt.Sprintf("%s/.story/story/config/config.toml", homeDir)
	// Read the config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	configContent := string(data)
	// Update seeds and peers
	configContent = regexp.MustCompile(`(?m)^seeds *=.*`).ReplaceAllString(configContent, fmt.Sprintf(`seeds = "%s"`, seeds))
	configContent = regexp.MustCompile(`(?m)^persistent_peers *=.*`).ReplaceAllString(configContent, fmt.Sprintf(`persistent_peers = "%s"`, peers))

	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		return err
	}
	return nil
}

func downloadGenesisAndAddrbook(homeDir string) error {
	pterm.Info.Println("Downloading genesis and addrbook...")

	endpoint, err := config.FetchItrocketRootEndpointFromAPI()
	if err != nil {
		return err
	}

	err = bash.RunCommand("wget", "-q", "-O", homeDir+"/.story/story/config/genesis.json", "https://"+endpoint+"/testnet/story/genesis.json")
	if err != nil {
		return err
	}

	err = bash.RunCommand("wget", "-q", "-O", homeDir+"/.story/story/config/addrbook.json", "https://"+endpoint+"/testnet/story/addrbook.json")
	if err != nil {
		return err
	}
	return nil
}

func createServiceFilesWithoutCosmovisor(homeDir, customPort string) error {
	// Create geth service file
	err := createGethServiceFile(homeDir, customPort)
	if err != nil {
		return err
	}

	// Create story service file
	storyServiceContent := fmt.Sprintf(`[Unit]
Description=Story Service
After=network.target

[Service]
User=%s
WorkingDirectory=%s/.story/story
ExecStart=%s/go/bin/story run

Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`, os.Getenv("USER"), homeDir, homeDir)

	err = os.WriteFile("/etc/systemd/system/story.service", []byte(storyServiceContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func createServiceFilesWithCosmovisor(homeDir, customPort string) error {
	// Create geth service file
	err := createGethServiceFile(homeDir, customPort)
	if err != nil {
		return err
	}

	// Create story service file
	storyServiceContent := fmt.Sprintf(`[Unit]
Description=Story Consensus Client
After=network.target

[Service]
User=%s
Environment="DAEMON_NAME=story"
Environment="DAEMON_HOME=%s/.story/story"
Environment="DAEMON_ALLOW_DOWNLOAD_BINARIES=true"
Environment="DAEMON_RESTART_AFTER_UPGRADE=true"
Environment="DAEMON_DATA_BACKUP_DIR=%s/.story/story/data"
Environment="UNSAFE_SKIP_BACKUP=true"
ExecStart=%s/go/bin/cosmovisor run run
Restart=always
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`, os.Getenv("USER"), homeDir, homeDir, homeDir)

	err = os.WriteFile("/etc/systemd/system/story.service", []byte(storyServiceContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func createGethServiceFile(homeDir, customPort string) error {
	// Create geth service file
	gethServiceContent := fmt.Sprintf(`[Unit]
Description=Story Geth daemon
After=network-online.target

[Service]
User=%s
ExecStart=%s/go/bin/story-geth --odyssey --syncmode full --http --http.api eth,net,web3,engine --http.vhosts '*' --http.addr 0.0.0.0 --http.port %s545 --authrpc.port %s551 --ws --ws.api eth,web3,net,txpool --ws.addr 0.0.0.0 --ws.port %s546
Restart=on-failure
RestartSec=3
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
`, os.Getenv("USER"), homeDir, customPort, customPort, customPort)

	err := os.WriteFile("/etc/systemd/system/story-geth.service", []byte(gethServiceContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func replaceInFile(filePath, old, new string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := string(data)
	re := regexp.MustCompile(old)
	content = re.ReplaceAllString(content, new)
	return os.WriteFile(filePath, []byte(content), 0644)
}

func getPublicIP() (string, error) {
	cmd := exec.Command("wget", "-qO-", "eth0.me")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getLatestBinaryVersions(binary string) (string, error) {
	apiURL := "https://snapshot-external-providers-api.krews.xyz/story/latest_versions"
	resp, err := http.Get(apiURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest versions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-OK status code %d from API", resp.StatusCode)
	}
	var version string
	var versions map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", fmt.Errorf("failed to decode API response: %v", err)
	}

	if binary == "geth" {
		gethVersion, exists := versions["geth-version"]
		if !exists {
			return "", fmt.Errorf("geth-version key not found in API response")
		}
		version = gethVersion

	} else if binary == "story" {
		storyVersion, exists := versions["story-version"]
		if !exists {
			return "", fmt.Errorf("geth-version key not found in API response")
		}
		version = storyVersion
	} else {
		return "", fmt.Errorf("no valid binaries provided")
	}

	return version, nil
}

func downloadBinary(binary string, homeDir string) error {
	binaryVersion, err := getLatestBinaryVersions(binary)
	if err != nil {
		return fmt.Errorf("failed to fetch "+binary+" version: %v", err)
	}

	if binary == "geth" {
		pterm.Info.Println("Downloading geth binary...")

		downloadURL := fmt.Sprintf("https://github.com/piplabs/story-geth/releases/download/%s/geth-linux-amd64", binaryVersion)

		err = bash.RunCommand("wget", "-O", "story-geth", downloadURL)
		if err != nil {
			return fmt.Errorf("failed to download Geth binary: %v", err)
		}

		// Make geth executable
		pterm.Info.Println("Setting execute permissions for story-geth...")
		err = bash.RunCommand("chmod", "+x", "story-geth")
		if err != nil {
			return err
		}

		// Move geth to ~/go/bin/
		pterm.Info.Println("Moving story-geth to " + homeDir + "/go/bin/")
		err = bash.RunCommand("rm", "-rf", homeDir+"/go/bin/story-geth")
		if err != nil {
			return err
		}
		err = bash.RunCommand("mkdir", "-p", homeDir+"/go/bin")
		if err != nil {
			return err
		}
		err = bash.RunCommand("mv", homeDir+"/story-geth", homeDir+"/go/bin/")
		if err != nil {
			return err
		}

	} else if binary == "story" {
		pterm.Info.Println("Downloading story binary...")

		downloadURL := fmt.Sprintf("https://github.com/piplabs/story/releases/download/%s/story-linux-amd64", binaryVersion)

		err = bash.RunCommand("wget", "-O", "story", downloadURL)
		if err != nil {
			return fmt.Errorf("failed to download Geth binary: %v", err)
		}

		// Move story binary to ~/go/bin/
		pterm.Info.Println("Moving story binary to " + homeDir + "/go/bin/")
		err = bash.RunCommand("rm", "-rf", homeDir+"/go/bin/story")
		if err != nil {
			return err
		}

		err = bash.RunCommand("chmod", "+x", homeDir+"/story")
		if err != nil {
			return err
		}

		err = bash.RunCommand("mv", homeDir+"/story", homeDir+"/go/bin/")
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("no valid binaries provided")
	}

	return nil
}

// setupWithCosmovisor sets up a Story node with Cosmovisor
func setupWithCosmovisor(moniker, customPort, pruningMode string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	os.Setenv("DAEMON_NAME", "story")
	os.Setenv("DAEMON_HOME", fmt.Sprintf("%s/.story/story", homeDir))

	// Add environment variables to .bash_profile
	bashProfile := fmt.Sprintf("%s/.bash_profile", homeDir)
	bashVars := []struct {
		Key   string
		Value string
	}{
		{"DAEMON_NAME", "story"},
		{"DAEMON_HOME", fmt.Sprintf("%s/.story/story", homeDir)},
	}

	for _, env := range bashVars {
		exportLine := fmt.Sprintf("export %s=%s", env.Key, env.Value)
		if err := ensureLineInFile(bashProfile, exportLine); err != nil {
			return fmt.Errorf("failed to update .bash_profile: %v", err)
		}
	}

	// Install Cosmovisor
	pterm.Info.Println("Installing Cosmovisor...")
	if err := bash.RunCommand("env", "PATH=$PATH:/usr/local/go/bin:$HOME/go/bin", "go", "install", "cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@latest"); err != nil {
		return fmt.Errorf("failed to install Cosmovisor: %v", err)
	}

	pterm.Success.Println("Cosmovisor installed successfully.")

	// Initialize Cosmovisor directories
	pterm.Info.Println("Initializing Cosmovisor directories...")

	latestBinaryVersion, err := getLatestBinaryVersions("story")
	if err != nil {
		return err
	}

	var upgradesDir = fmt.Sprintf("%s/.story/story", homeDir) + fmt.Sprintf("/cosmovisor/upgrades/%s/bin", latestBinaryVersion)
	if err := bash.RunCommand("mkdir", "-p", upgradesDir); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", upgradesDir, err)
	}

	// Download get Binary
	downloadBinary("geth", homeDir)

	// Download and configure binaries
	downloadBinary("story", homeDir)

	if err := bash.RunCommand(homeDir+"/go/bin/cosmovisor", "init", homeDir+"/go/bin/story"); err != nil {
		return fmt.Errorf("failed to copy binary to upgrades directory: %v", err)
	}

	if err := bash.RunCommand("cp", homeDir+"/go/bin/story", upgradesDir+"/story"); err != nil {
		return fmt.Errorf("failed to copy binary to upgrades directory: %v", err)
	}

	// Initialize Story
	initializingStoryNode(homeDir, moniker)

	// Configure seeds and peers
	err = configureSeedsAndPeers(homeDir)
	if err != nil {
		return err
	}

	// Download genesis and addrbook
	err = downloadGenesisAndAddrbook(homeDir)
	if err != nil {
		return err
	}

	// Set config.toml path
	configToml := fmt.Sprintf("%s/.story/story/config/config.toml", homeDir)

	// Set custom ports in story.toml
	settingCustomPorts(homeDir, configToml)

	// Set indexer mode if pruned
	setIndexer(configToml, pruningMode)

	// Update systemd service file
	createServiceFilesWithCosmovisor(homeDir, customPort)

	// Enable and start services
	pterm.Info.Println("Enabling services...")
	if err := bash.RunCommand("sudo", "systemctl", "daemon-reload"); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}
	if err := bash.RunCommand("sudo", "systemctl", "enable", "story", "story-geth"); err != nil {
		return fmt.Errorf("failed to enable story service: %v", err)
	}

	// Download snapshot based on provider
	pterm.Info.Println("Downloading snapshot...")
	snapshot.CallRunDownloadSnapshotManually(pruningMode, homeDir, true)

	pterm.Success.Println("Cosmovisor setup completed successfully.")

	// Inform the user to source the .bash_profile manually
	coloredText := pterm.FgRed.Sprint("source ~/.bash_profile")
	coloredRestart := pterm.FgRed.Sprint("scli restart --all")
	pterm.Warning.Println("Please run", coloredText, "or restart your terminal to apply the environment variables.")
	pterm.Warning.Println("Then run the", coloredRestart, "command to start cosmovisor.")

	return nil
}

// ensureLineInFile checks if a line exists in a file, and appends it if not
func ensureLineInFile(filename, line string) error {
	// Read the file
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, create it
			return appendToFile(filename, line)
		}
		return err
	}

	content := string(data)
	if !strings.Contains(content, line) {
		return appendToFile(filename, line)
	}
	return nil
}

// appendToFile appends a single line to a file
func appendToFile(filename, line string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(line + "\n"); err != nil {
		return err
	}
	return nil
}

func createNecessaryDirectories() error {
	// Create necessary directories
	pterm.Info.Println("Creating necessary directories...")
	err := bash.RunCommand("bash", "-c", "[ ! -d \"$HOME/.story/story\" ] && mkdir -p \"$HOME/.story/story\"")
	if err != nil {
		return err
	}

	return nil
}

func initializingStoryNode(homeDir string, moniker string) error {

	pterm.Info.Println("Initializing Story node...")
	err := bash.RunCommand(homeDir+"/go/bin/story", "init", "--moniker", moniker, "--network", "odyssey")
	if err != nil {
		return err
	}

	return nil
}

func settingCustomPorts(homeDir string, configToml string) error {
	pterm.Info.Println("Setting custom ports in story.toml...")
	storyToml := fmt.Sprintf("%s/.story/story/config/story.toml", homeDir)
	err := replaceInFile(storyToml, `:1317`, fmt.Sprintf(":%s317", customPort))
	if err != nil {
		return err
	}
	err = replaceInFile(storyToml, `:8551`, fmt.Sprintf(":%s551", customPort))
	if err != nil {
		return err
	}

	// Set custom ports in config.toml
	pterm.Info.Println("Setting custom ports in config.toml...")
	publicIP, err := getPublicIP()
	if err != nil {
		return err
	}
	externalAddress := fmt.Sprintf("external_address = \"%s:%s656\"", publicIP, customPort)
	err = replaceInFile(configToml, `:26658`, fmt.Sprintf(":%s658", customPort))
	if err != nil {
		return err
	}
	err = replaceInFile(configToml, `:26657`, fmt.Sprintf(":%s657", customPort))
	if err != nil {
		return err
	}
	err = replaceInFile(configToml, `:26656`, fmt.Sprintf(":%s656", customPort))
	if err != nil {
		return err
	}
	err = replaceInFile(configToml, `^external_address = .*`, externalAddress)
	if err != nil {
		return err
	}
	err = replaceInFile(configToml, `:26660`, fmt.Sprintf(":%s660", customPort))
	if err != nil {
		return err
	}
	return nil
}

func enablePrometheus(configToml string) error {
	pterm.Info.Println("Enabling Prometheus...")
	err := replaceInFile(configToml, "prometheus = false", "prometheus = true")
	if err != nil {
		return err
	}

	return nil
}

func setIndexer(configToml string, pruningMode string) error {
	if strings.ToLower(pruningMode) == "pruned" {
		err := replaceInFile(configToml, `^indexer *=.*`, `indexer = "null"`)
		if err != nil {
			return err
		}
	}
	return nil
}

func selectSetupType() (string, error) {
	if setupType == "" {
		pmPrompt := promptui.Select{
			Label: "Select setup type",
			Items: []string{"without cosmovisor", "with cosmovisor"},
		}
		_, pmResult, err := pmPrompt.Run()
		if err != nil {
			return "", err
		}
		setupType = pmResult
	}
	return setupType, nil
}

package cmd

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/pierrec/lz4/v4"
	"github.com/pterm/pterm"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/spf13/cobra"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

var (
	moniker          string
	customPort       string
	snapshotProvider string
)

var (
	prunedSnapshotSize  string
	archiveSnapshotSize string
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

// Helper functions for colored messages
func printInfo(message string) {
	pterm.Info.Printf("%s\n", message)
}

func printWarning(message string) {
	pterm.Warning.Printf("%s\n", message)
}

func printSuccess(message string) {
	pterm.Success.Printf("%s\n", message)
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
			err := os.RemoveAll(storyDir)
			if err != nil {
				return err
			}
			err = runCommand("systemctl stop story story-geth")
			if err != nil {
				return err
			}
			printSuccess("Existing installation removed.")
		} else {
			printWarning("Setup aborted.")
			return nil
		}
	}

	// Step 2: Fetch Snapshot Sizes from API (only for Krews)
	//err = fetchSnapshotSizes()
	//if err != nil {
	//	printWarning(fmt.Sprintf("Failed to fetch snapshot sizes: %v", err))
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

	// Pruning mode info message
	printInfo("\nPruning Mode Information:")
	fmt.Printf(" - Pruned Mode: Stores only recent blockchain data, reducing disk usage. Snapshot size: %s\n", prunedSnapshotSize)
	fmt.Printf(" - Archive Mode: Stores the entire blockchain history, requiring more disk space. Snapshot size: %s\n\n", archiveSnapshotSize)

	if pruningMode == "" {
		prompt := promptui.Select{
			Label: "Select the pruning mode",
			Items: []string{"pruned", "archive"},
		}
		_, pruningMode, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	// Step 4: Select Snapshot Provider
	prompt := promptui.Select{
		Label: "Select the snapshot provider",
		Items: []string{"Itrocket", "Krews"},
	}
	_, snapshotProvider, err = prompt.Run()
	if err != nil {
		return err
	}

	// Proceed with setup without Cosmovisor
	err = setupWithoutCosmovisor(moniker, customPort, pruningMode, snapshotProvider)
	if err != nil {
		return err
	}

	return nil
}

func checkSystemResources() error {
	printInfo("Checking system resources...")

	// CPU cores
	cpuCores := runtime.NumCPU()
	if cpuCores < recommendedCPU {
		printWarning(fmt.Sprintf("You have %d CPU cores. Recommended is %d cores.", cpuCores, recommendedCPU))
	} else {
		printInfo(fmt.Sprintf("CPU cores: %d", cpuCores))
	}

	// RAM
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return err
	}
	ramMB := vmStat.Total / (1024 * 1024)
	if int(ramMB) < recommendedRAM {
		printWarning(fmt.Sprintf("You have %d MB of RAM. Recommended is %d MB.", ramMB, recommendedRAM))
	} else {
		printInfo(fmt.Sprintf("RAM: %d MB", ramMB))
	}

	// Disk space
	diskStat, err := disk.Usage("/")
	if err != nil {
		return err
	}
	diskGB := diskStat.Total / (1024 * 1024 * 1024)
	recommendedDiskGB := recommendedDisk / (1024 * 1024 * 1024)
	if diskStat.Total < uint64(recommendedDisk) {
		printWarning(fmt.Sprintf("You have %d GB of disk space. Recommended is %d GB.", diskGB, recommendedDiskGB))
	} else {
		printInfo(fmt.Sprintf("Disk space: %d GB", diskGB))
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

func setupWithoutCosmovisor(moniker, customPort, pruningMode, snapshotProvider string) error {
	homeDir, _ := os.UserHomeDir()
	os.Setenv("MONIKER", moniker)
	os.Setenv("STORY_PORT", customPort)
	os.Setenv("PRUNING_MODE", pruningMode)

	// Navigate to home directory
	printInfo("Navigating to home directory...")
	err := runCommand("cd $HOME")
	if err != nil {
		return err
	}

	// Download geth binary
	printInfo("Downloading geth binary...")
	err = runCommand("wget -O geth https://github.com/piplabs/story-geth/releases/latest/download/geth-linux-amd64")
	if err != nil {
		return err
	}

	// Make geth executable
	printInfo("Setting execute permissions for geth...")
	err = runCommand("chmod +x geth")
	if err != nil {
		return err
	}

	// Move geth to ~/go/bin/
	printInfo("Moving geth to ~/go/bin/")
	err = runCommand("rm -rf ~/go/bin/geth")
	if err != nil {
		return err
	}
	err = runCommand("mkdir -p \"$HOME/go/bin\"")
	if err != nil {
		return err
	}
	err = runCommand("mv $HOME/geth $HOME/go/bin/")
	if err != nil {
		return err
	}

	// Create necessary directories
	printInfo("Creating necessary directories...")
	err = runCommand("[ ! -d \"$HOME/.story/story\" ] && mkdir -p \"$HOME/.story/story\"")
	if err != nil {
		return err
	}
	err = runCommand("[ ! -d \"$HOME/.story/geth\" ] && mkdir -p \"$HOME/.story/geth\"")
	if err != nil {
		return err
	}

	// Install Story
	printInfo("Cloning Story repository...")
	err = runCommand("cd $HOME && rm -rf story && git clone https://github.com/piplabs/story")
	if err != nil {
		return err
	}

	// Get the latest release tag
	tag, err := getLatestReleaseTag("piplabs/story")
	if err != nil {
		return err
	}

	printInfo(fmt.Sprintf("Checking out version %s...", tag))
	err = runCommand(fmt.Sprintf("cd story && git checkout %s", tag))
	if err != nil {
		return err
	}

	printInfo("Building Story binary...")
	err = runCommand("cd story && go build -o story ./client")
	if err != nil {
		return err
	}

	// Move story binary to ~/go/bin/
	printInfo("Moving story binary to ~/go/bin/")
	err = runCommand("rm -rf $HOME/go/bin/story")
	if err != nil {
		return err
	}
	err = runCommand("mv $HOME/story/story $HOME/go/bin/")
	if err != nil {
		return err
	}

	// Initialize Story
	printInfo("Initializing Story node...")
	initCmd := fmt.Sprintf("story init --moniker %s --network iliad", moniker)
	err = runCommand(initCmd)
	if err != nil {
		return err
	}

	// Configure seeds and peers
	printInfo("Configuring seeds and peers...")
	err = configureSeedsAndPeersWithoutCosmovisor(homeDir)
	if err != nil {
		return err
	}

	// Download genesis and addrbook
	printInfo("Downloading genesis and addrbook...")
	err = downloadGenesisAndAddrbookWithoutCosmovisor(homeDir)
	if err != nil {
		return err
	}

	// Set custom ports in story.toml
	printInfo("Setting custom ports in story.toml...")
	storyToml := fmt.Sprintf("%s/.story/story/config/story.toml", homeDir)
	err = replaceInFile(storyToml, `:1317`, fmt.Sprintf(":%s317", customPort))
	if err != nil {
		return err
	}
	err = replaceInFile(storyToml, `:8551`, fmt.Sprintf(":%s551", customPort))
	if err != nil {
		return err
	}

	// Set custom ports in config.toml
	printInfo("Setting custom ports in config.toml...")
	configToml := fmt.Sprintf("%s/.story/story/config/config.toml", homeDir)
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

	// Enable Prometheus
	printInfo("Enabling Prometheus...")
	err = replaceInFile(configToml, "prometheus = false", "prometheus = true")
	if err != nil {
		return err
	}

	// Set pruning mode if pruned
	if strings.ToLower(pruningMode) == "pruned" {
		err = replaceInFile(configToml, `^indexer *=.*`, `indexer = "null"`)
		if err != nil {
			return err
		}
	}

	// Create systemd service files
	printInfo("Creating systemd service files...")
	err = createServiceFilesWithoutCosmovisor(homeDir, customPort)
	if err != nil {
		return err
	}

	// Download snapshot based on provider
	printInfo("Downloading snapshot...")
	if strings.ToLower(snapshotProvider) == "krews" {
		err = downloadSnapshotKrews(homeDir, pruningMode)
	} else if strings.ToLower(snapshotProvider) == "itrocket" {
		err = downloadSnapshotItrocket(homeDir, pruningMode)
	} else {
		return errors.New("unknown snapshot provider")
	}
	if err != nil {
		return err
	}

	// Enable and start services
	printInfo("Enabling and starting services...")
	err = runCommand("sudo systemctl daemon-reload")
	if err != nil {
		return err
	}
	err = runCommand("sudo systemctl enable story story-geth")
	if err != nil {
		return err
	}
	err = runCommand("sudo systemctl restart story story-geth")
	if err != nil {
		return err
	}
	printSuccess("Node setup without Cosmovisor completed successfully.")

	return nil
}

func installAndConfigureRcloneKrews(homeDir string) error {
	// Check if Rclone is installed
	_, err := exec.LookPath("rclone")
	if err != nil {
		printInfo("Rclone not found. Installing Rclone...")
		// Install Rclone
		installCmd := "sudo -v ; curl https://rclone.org/install.sh | sudo bash"
		err = runCommand(installCmd)
		if err != nil {
			return err
		}
		printSuccess("Rclone Installed")
	} else {
		printInfo("Rclone is Already Installed")
	}

	// Configure Rclone for Krews
	printInfo("Configuring Rclone for Krews...")
	rcloneConf := `[krews-snapshot]
type = s3
provider = DigitalOcean
region = fra1
endpoint = https://fra1.cdn.digitaloceanspaces.com
`
	rcloneConfDir := fmt.Sprintf("%s/.config/rclone", homeDir)
	err = os.MkdirAll(rcloneConfDir, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s/rclone.conf", rcloneConfDir), []byte(rcloneConf), 0644)
	if err != nil {
		return err
	}

	return nil
}

// Helper function to download a file with a progress bar
func downloadFileWithProgress(url, dest string) error {
	// Create the file
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Get the size
	sizeStr := resp.Header.Get("Content-Length")
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size <= 0 {
		return errors.New("invalid Content-Length")
	}

	// Create progress bar
	p := mpb.New(
		mpb.WithWidth(64),
		mpb.WithRefreshRate(180*time.Millisecond),
	)
	bar := p.AddBar(int64(size),
		mpb.PrependDecorators(
			decor.Name("Downloading:", decor.WC{W: len("Downloading: "), C: decor.DidentRight}),
			decor.CountersKibiByte("% .2f / % .2f"),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WC{W: 5}),
		),
	)

	// Create a proxy reader
	reader := bar.ProxyReader(resp.Body)
	defer reader.Close()

	// Write the body to file
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}

	p.Wait()

	return nil
}

// Helper function to decompress LZ4 and extract tar
func decompressAndExtractLz4Tar(lz4Path, destDir string) error {
	// Open the LZ4 file
	file, err := os.Open(lz4Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create LZ4 reader
	lz4Reader := lz4.NewReader(file)

	// Create tar reader
	tarReader := tar.NewReader(lz4Reader)

	// Iterate through the files in the tar archive
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return err
		}

		// Determine proper file path
		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			// Create file
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			// Set permissions
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		default:
			// Unsupported type
			printWarning(fmt.Sprintf("Unsupported file type: %v in %s", header.Typeflag, header.Name))
		}
	}

	return nil
}

func configureSeedsAndPeersWithoutCosmovisor(homeDir string) error {
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

func downloadGenesisAndAddrbookWithoutCosmovisor(homeDir string) error {
	cmds := []string{
		fmt.Sprintf("wget -q -O %s/.story/story/config/genesis.json https://server-3.itrocket.net/testnet/story/genesis.json", homeDir),
		fmt.Sprintf("wget -q -O %s/.story/story/config/addrbook.json https://server-3.itrocket.net/testnet/story/addrbook.json", homeDir),
	}
	for _, cmdStr := range cmds {
		err := runCommand(cmdStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func createServiceFilesWithoutCosmovisor(homeDir, customPort string) error {
	// Create geth service file
	gethServiceContent := fmt.Sprintf(`[Unit]
Description=Story Geth daemon
After=network-online.target

[Service]
User=%s
ExecStart=%s/go/bin/geth --iliad --syncmode full --http --http.api eth,net,web3,engine --http.vhosts '*' --http.addr 0.0.0.0 --http.port %s545 --authrpc.port %s551 --ws --ws.api eth,web3,net,txpool --ws.addr 0.0.0.0 --ws.port %s546
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

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n✖ Error executing command: %s %v\n", name, args)
		fmt.Println("Error:", err)
		fmt.Println("Stdout:", stdout.String())
		fmt.Println("Stderr:", stderr.String())
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

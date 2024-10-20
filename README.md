
# StoryCLI

`StoryCLI` is a command line interface designed to manage various operations related to Story Protocol node management. The tool includes commands for setting up nodes, starting, stopping and configuring services, as well as managing snapshots and logs.

## Installation

To install StoryCLI, you can use the following bash command:

```bash
sudo -v ; curl https://krews-eu.krews-storage.xyz/install.sh | sudo bash
```

### Prerequisites

Ensure that you have Go installed before running the installation script. The installation script will automatically install go if it is not installed. You can check if Go is installed using:

```bash
go version
```

## Usage

Once installed, the following commands are available in StoryCLI:

### Root Command

This is the base command. You can list all available commands by running:

```bash
scli
```

### Commands


#### `setup node`

Sets up an easy story node setup by asking you questions.

Usage:

```bash
scli setup node
```
example outout:

```bash
 INFO  Checking system resources...
 INFO  CPU cores: 4
 WARNING  You have 7941 MB of RAM. Recommended is 16384 MB.
 INFO  Disk space: 231 GB
✔ Yes
 SUCCESS  Existing installation removed.
 INFO  Fetching snapshot sizes from Krews...
Enter your moniker: test
Enter the first two digits of the custom port (default 26): : 23
 INFO
       Pruning Mode Information:
 - Pruned Mode: Stores only recent blockchain data, reducing disk usage. Snapshot size: 121G
 - Archive Mode: Stores the entire blockchain history, requiring more disk space. Snapshot size: 495G

✔ pruned
✔ Krews
 INFO  Navigating to home directory...
 INFO  Downloading geth binary...
 INFO  Setting execute permissions for geth...
 INFO  Moving geth to ~/go/bin/
 INFO  Creating necessary directories...
 INFO  Cloning Story repository...
 INFO  Checking out version v0.11.0...
 INFO  Building Story binary...
 INFO  Moving story binary to ~/go/bin/
 INFO  Initializing Story node...
 INFO  Configuring seeds and peers...
 INFO  Downloading genesis and addrbook...
 INFO  Setting custom ports in story.toml...
 INFO  Setting custom ports in config.toml...
 INFO  Enabling Prometheus...
 INFO  Creating systemd service files...
 INFO  Downloading snapshot...
 INFO  Installing required packages for Krews snapshot...
 INFO  Stopping Story and Story-Geth services...
 INFO  Backup priv_validator_state.json...
 INFO  Removing old Story data and unpacking new snapshot...
 INFO  Downloading Story snapshot...
Downloading: 7.08 GiB / 7.08 GiB [==============================================================] 100 %
 INFO  Extracting Story snapshot...
 INFO  Removing old Geth data and downloading new Geth snapshot...
 INFO  Downloading Geth snapshot...
Downloading: 41.32 GiB / 41.32 GiB [==============================================================] 100 %
 INFO  extracting Geth snapshot...
 INFO  Restoring priv_validator_state.json...
 INFO  Starting Story and Story-Geth services...
 SUCCESS  Snapshot successfully downloaded and applied from Krews.
 INFO  Enabling and starting services...
 SUCCESS  Node setup without Cosmovisor completed successfully.

```

#### `logs`

This command retrieves and displays story and geth logs from the services.

Usage:

```bash
scli logs [service]
```

#### `restart`

Restarts Story node. Commonly used to refresh the system after changes or errors.

Usage:

```bash
scli restart
```

#### `set`

This command allows the user to set configurations for the story node.

Usage:

```bash
scli set [configuration] [value]
```

#### `show`

Displays current story configuration parameters.

Usage:

```bash
scli show [configuration param]
```

#### `snapshot`

Downloads snapshots from a snapshot provider and installs them on Story node data

Usage:

```bash
scli snapshot
```
example output:

```bash
 INFO  Fetching snapshot sizes from Krews...
 INFO
       Pruning Mode Information:
 - Pruned Mode: Stores only recent blockchain data, reducing disk usage. Snapshot size: 121G
 - Archive Mode: Stores the entire blockchain history, requiring more disk space. Snapshot size: 496G

✔ pruned
✔ Krews
 INFO  stopping services...
 INFO  Fetching snapshot names from Krews...
 INFO  Installing required packages for Krews snapshot...
 INFO  Stopping Story and Story-Geth services...
 INFO  Backup priv_validator_state.json...
 INFO  Removing old Story data...
 INFO  Downloading Story snapshot...
Downloading: 7.08 GiB / 7.08 GiB [==============================================================] 100 %
 INFO  Extracting Story snapshot...
 INFO  Removing old Geth data and downloading new Geth snapshot...
 INFO  Downloading Geth snapshot...
Downloading: 41.32 GiB / 41.32 GiB [==============================================================] 100 %
 INFO  extracting Geth snapshot...
 INFO  Restoring priv_validator_state.json...
 INFO  Starting Story and Story-Geth services...
 SUCCESS  Snapshot successfully downloaded and applied from Krews.
 INFO  run scli restart to restart the systemd services and apply the new values

```

#### `status`

Checks the status of Story and Geth services.

Usage:

```bash
scli status
```

#### `stop`

Stops the running Story and Geth services.

Usage:

```bash
scli stop
```

#### `update`

Updates Story and Geth binaries.

Usage:

```bash
scli update
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

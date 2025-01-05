# StoryCLI

`StoryCLI` is a command line interface designed to manage various operations related to Story Protocol node management. The tool includes commands for setting up nodes, starting, stopping and configuring services, as well as managing snapshots and logs.

## Installation

To install StoryCLI, you can use the following bash command:

```bash
sudo -v ; curl https://krews-eu.krews-storage.xyz/install.sh | sudo bash
```

### Prerequisites

Ensure that you have Go installed before running the installation script. The installation script will automatically install Go if it is not installed. You can check if Go is installed using:

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

Sets up an easy Story node setup by asking you questions. This includes options for `pruned` and `archive` pruning modes. The command also supports setting up the node with or without Cosmovisor.

Usage:

```bash
scli setup node
```

Example output:

```bash
 INFO  Checking system resources...
 INFO  CPU cores: 8
 WARNING  You have 15988 MB of RAM. Recommended is 16384 MB.
 INFO  Disk space: 309 GB
✔ Yes
 INFO  Checking if Story and Story-Geth services are active...
 INFO  Stopping Story services...
 SUCCESS  Existing installation removed.
Enter your moniker: test
Enter the first two digits of the custom port (default 26): : 23
✔ without cosmovisor
 INFO  Pruning Mode Information:
 INFO   - Pruned Mode: Stores only recent blocks, reducing disk usage.
 INFO   - Archive Mode: Stores the entire chain history, requiring more disk space.
✔ pruned
 INFO  Navigating to home directory...
 INFO  Downloading geth binary...
 INFO  Setting execute permissions for story-geth...
 INFO  Moving story-geth to /root/go/bin/
 INFO  Creating necessary directories...
 INFO  Downloading story binary...
 INFO  Moving story binary to /root/go/bin/
 INFO  Initializing Story node...
 INFO  Configuring seeds and peers...
 INFO  Downloading genesis and addrbook...
 INFO  Setting custom ports in story.toml...
 INFO  Setting custom ports in config.toml...
 INFO  Enabling Prometheus...
 INFO  Creating systemd service files...
 INFO  Enabling services...
 INFO  Downloading snapshot...
 INFO  Fetching snapshot data for providers (mode=pruned)...
Jnode - ( mode: pruned | size: 58.92G | height: 1770319 | 41m )
 INFO  Installing required packages for Jnode snapshot...
 INFO  Stopping Story and Story-Geth services...
 INFO  Backing up priv_validator_state.json...
 INFO  Removing old Story and Geth data...
 INFO  Downloading Story snapshot...
[4.1GiB/4.1GiB(100%) CN:10 DL:243MiB]ETA:1s]
 INFO  Download complete!
 INFO  Downloading Geth snapshot...
[54GiB/54GiB(100%) CN:2 DL:102MiB]]ETA:1s]
 INFO  Download complete!
 INFO  Extracting Story snapshot...
 INFO  Extracting Geth snapshot...
 INFO  Restoring priv_validator_state.json...
 INFO  Starting Story and Story-Geth services...
 SUCCESS  Snapshot successfully downloaded and applied from Jnode.
 SUCCESS  Node setup without Cosmovisor completed successfully.
```

#### `snapshot providers`

Lists available snapshot providers and displays their data in a table format. Separate tables are created for `pruned` and `archive` modes.

Usage:

```bash
scli snapshot providers
```

Example output:

```bash
# Pruned Snapshots

Provider   | Total Size | Block Height | Time Ago
Itrocket   | 63.70G     | 1773764      | 1h 4m ago
Krews      | 197G       | 1773702      | 37m ago
Jnode      | 58.92G     | 1770319      | 4h 39m ago
Mandragora | 1.24M      | unknown      | unknown

# Archive Snapshots

Provider   | Total Size | Block Height | Time Ago
Itrocket   | 265.00G    | 1768887      | 6h 14m ago
Krews      | 304G       | 1772449      | 2h 9m ago
Jnode      | 265.35G    | 1770327      | 4h 38m ago
Mandragora | 60.60G     | 1773606      | 1h 10m ago
```

#### `snapshot download`

Downloads snapshots from a selected provider and applies them to initialize or update your Story node.

Usage:

```bash
scli snapshot download
```

Example output:
```bash
 INFO  Pruning Mode Information:
 INFO   - Pruned Mode: Stores only recent blocks, reducing disk usage.
 INFO   - Archive Mode: Stores the entire chain history, requiring more disk space.
✔ pruned
 INFO  Fetching snapshot data for providers (mode=pruned)...
Use the arrow keys to navigate: ↓ ↑ → ← 
Select the snapshot provider
  ✔ Itrocket ( mode: pruned | size: 63.70G | height: 1773764 | 2h 17m ago )
    Jnode ( mode: pruned | size: 58.92G | height: 1770319 | 5h 53m )
    Mandragora ( mode: pruned | size: 1.24M | height: unknown | unknown )
    Krews ( mode: pruned | size: 197G | height: 1773702 | 1h 50m ago )
```
Flags:

- `--output-path`: Download snapshot directly to the specified path.
- `--home`: Specify the home directory for the Story node.

#### `logs`

Retrieves and displays logs for the Story and Geth services.

Usage:

```bash
scli logs [service]
```

Flags:

- `--lines`: Number of log lines to display (default: 20).

#### `restart`

Restarts Story and Story-Geth services. Supports restarting both services together or individually.

Usage:

```bash
scli restart
```

Flags:

- `--all`: Restart both services together.

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

#### `status`

Checks the status of Story and Story-Geth services.

Usage:

```bash
scli status
```

#### `stop`

Stops the running Story and Story-Geth services.

Usage:

```bash
scli stop
```

#### `update`

Updates Story and Geth binaries to the latest version.

Usage:

```bash
scli update
```


## Features

- **Snapshot Management:** Download and apply snapshots from multiple providers (Itrocket, Krews, Jnode, Mandragora).
- **Node Setup:** Streamlined setup for Story nodes, with options for `pruned` or `archive` modes.
- **Logs and Status:** View logs and check the status of Story and Story-Geth services.
- **Restart and Stop Commands:** Easily manage node services.
- **Custom Port and Cosmovisor Support:** Flexibility to configure ports and set up with or without Cosmovisor.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

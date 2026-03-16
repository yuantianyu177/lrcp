# lrcp

Bidirectional file transfer tool over SSH, powered by rsync and SSH ControlMaster.

## Features

- **Host management** — add, edit, remove, and list SSH hosts interactively
- **Persistent SSH connections** — uses ControlMaster to keep connections alive across transfers
- **Push/Pull** — bidirectional rsync-based file transfer with a simple CLI
- **Remote path completion** — tab-complete remote paths in push/pull commands
- **Password & key auth** — supports both SSH key and password authentication (encrypted with AES-256-GCM)
- **Configurable rsync** — customize rsync behavior via a config file
- **Port forwarding** — SSH tunnel with `lrcp tunnel`, supports local and remote forwarding
- **Shell completion** — bash, zsh, and fish completion support

## Installation

### From release binary

Download the binary for your platform from the [releases](https://github.com/yuantianyu177/lrcp/releases) page, rename it to `lrcp`, and place it in your `$PATH`.

### From source

```bash
go install github.com/dislab/lrcp@latest
```

### Shell completion

```bash
# bash
source <(lrcp completion bash)

# zsh (add to ~/.zshrc)
autoload -U compinit && compinit
source <(lrcp completion zsh)

# fish
lrcp completion fish | source
```

## Usage

### Add a host

```bash
lrcp new
```

Interactively prompts for host name, hostname, user, port, and authentication method.

### List hosts

```bash
lrcp list
```

Shows all configured hosts with connection status.

### Connect / Close

```bash
lrcp connect <host>    # establish SSH connection
lrcp close <host>      # close connection
lrcp close             # close all connections
```

### Push files to remote

```bash
lrcp push <local> <host>:<remote> [-- rsync_flags...]
```

### Pull files from remote

```bash
lrcp pull <host>:<remote> <local> [-- rsync_flags...]
```

### Port forwarding (tunnel)

```bash
# Forward remote server:7890 to local :9900
lrcp tunnel server:7890 localhost:9900

# Forward local :5432 to remote server:3243
lrcp tunnel localhost:5432 server:3243
```

Use `localhost` or `127.0.0.1` to indicate the local machine. Active tunnels are shown in `lrcp list`.

### Edit / Remove a host

```bash
lrcp edit <host>
lrcp remove <host>
```

## Configuration

All configuration is stored in `~/.config/lrcp/`:

```
~/.config/lrcp/
├── hosts          # SSH host definitions
├── credentials    # encrypted passwords (AES-256-GCM)
├── config         # rsync configuration
└── sockets/       # SSH ControlMaster sockets & tunnel registry
    └── tunnels.json  # active tunnel records
```

### Hosts file

SSH-config-like format:

```
Host myserver
  HostName 192.168.1.100
  User admin
  Port 22
  IdentityFile ~/.ssh/id_rsa

Host dev
  HostName dev.example.com
  User deploy
  Port 2222
  Password yes
```

### Rsync config

`~/.config/lrcp/config` controls default rsync flags. Uncomment lines to enable:

```
# Boolean flags
archive-mode archive true
verbose-output verbose true
compression compress true
show-progress progress true

# Value flags
bandwidth-limit bwlimit 1000

# List flags (applied as multiple --flag=value)
exclude-patterns exclude .git/ .DS_Store node_modules/
```

CLI args after `--` take highest priority:

```bash
lrcp push ./src server:~/project/ -- --delete --verbose
```

### Password encryption

Passwords are encrypted using AES-256-GCM with a key derived from `/etc/machine-id` via SHA-256. Credentials are stored in `~/.config/lrcp/credentials` with `0600` permissions.

Password authentication requires `sshpass`:

```bash
sudo apt install sshpass
```

## Build

```bash
# build for current platform
go build -o lrcp .

# build release binaries for all platforms
bash build.sh
```

The build script produces minimal binaries (`-s -w -trimpath`) for:

- linux/amd64, linux/arm64, linux/arm
- darwin/amd64, darwin/arm64
- windows/amd64, windows/arm64

Output is in the `dist/` directory.

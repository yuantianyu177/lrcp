package ssh

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// TunnelDirection indicates the type of SSH port forwarding.
type TunnelDirection int

const (
	// LocalForward forwards a remote port to local (-L).
	LocalForward TunnelDirection = iota
	// RemoteForward forwards a local port to remote (-R).
	RemoteForward
)

// TunnelEntry represents an active tunnel record.
type TunnelEntry struct {
	Host       string          `json:"host"`
	Direction  TunnelDirection `json:"direction"`
	From       string          `json:"from"`
	To         string          `json:"to"`
	BindAddr   string          `json:"bind_addr"`
	TargetAddr string          `json:"target_addr"`
	CreatedAt  time.Time       `json:"created_at"`
}

// TunnelOptions holds the parameters for SSH port forwarding.
type TunnelOptions struct {
	SocketPath string
	Direction  TunnelDirection
	BindAddr   string // bind address and port (e.g. "localhost:9900")
	TargetAddr string // target address and port (e.g. "localhost:7890")
}

// Tunnel creates a port forward via the existing ControlMaster socket using -O forward.
func Tunnel(opts TunnelOptions) error {
	var flag string
	switch opts.Direction {
	case LocalForward:
		flag = "-L"
	case RemoteForward:
		flag = "-R"
	}

	spec := fmt.Sprintf("%s:%s", opts.BindAddr, opts.TargetAddr)
	cmd := exec.Command("ssh",
		"-O", "forward",
		"-o", fmt.Sprintf("ControlPath=%s", opts.SocketPath),
		flag, spec,
		"_",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("forward failed: %w", err)
	}
	return nil
}

// CancelTunnel removes a port forward from the ControlMaster socket.
func CancelTunnel(socketPath string, direction TunnelDirection, bindAddr, targetAddr string) error {
	var flag string
	switch direction {
	case LocalForward:
		flag = "-L"
	case RemoteForward:
		flag = "-R"
	}

	spec := fmt.Sprintf("%s:%s", bindAddr, targetAddr)
	cmd := exec.Command("ssh",
		"-O", "cancel",
		"-o", fmt.Sprintf("ControlPath=%s", socketPath),
		flag, spec,
		"_",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cancel forward failed: %w", err)
	}
	return nil
}

// tunnelsFilePath returns the path to the tunnels registry file.
func tunnelsFilePath(socketsDir string) string {
	return filepath.Join(socketsDir, "tunnels.json")
}

// LoadTunnels reads the tunnel registry from disk.
func LoadTunnels(socketsDir string) ([]TunnelEntry, error) {
	data, err := os.ReadFile(tunnelsFilePath(socketsDir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []TunnelEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// SaveTunnels writes the tunnel registry to disk.
func SaveTunnels(socketsDir string, entries []TunnelEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tunnelsFilePath(socketsDir), data, 0600)
}

// AddTunnel appends a tunnel entry and saves.
func AddTunnel(socketsDir string, entry TunnelEntry) error {
	entries, err := LoadTunnels(socketsDir)
	if err != nil {
		return err
	}
	entries = append(entries, entry)
	return SaveTunnels(socketsDir, entries)
}

// RemoveTunnel removes a matching tunnel entry and saves.
func RemoveTunnel(socketsDir string, from, to string) error {
	entries, err := LoadTunnels(socketsDir)
	if err != nil {
		return err
	}
	var remaining []TunnelEntry
	for _, e := range entries {
		if e.From == from && e.To == to {
			continue
		}
		remaining = append(remaining, e)
	}
	return SaveTunnels(socketsDir, remaining)
}

// CleanTunnels removes entries whose ControlMaster socket is no longer active.
func CleanTunnels(socketsDir string) error {
	entries, err := LoadTunnels(socketsDir)
	if err != nil {
		return err
	}
	var alive []TunnelEntry
	for _, e := range entries {
		sockPath := filepath.Join(socketsDir, e.Host+".sock")
		if Check(sockPath) {
			alive = append(alive, e)
		}
	}
	return SaveTunnels(socketsDir, alive)
}

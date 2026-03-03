package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Paths holds all config directory paths.
type Paths struct {
	Dir        string // ~/.config/lrcp
	HostsFile  string // ~/.config/lrcp/hosts
	CredsFile  string // ~/.config/lrcp/credentials
	SocketsDir string // ~/.config/lrcp/sockets
	ConfigFile string // ~/.config/lrcp/config
}

// GetPaths returns the default config paths under ~/.config/lrcp.
func GetPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(home, ".config", "lrcp")
	return &Paths{
		Dir:        dir,
		HostsFile:  filepath.Join(dir, "hosts"),
		CredsFile:  filepath.Join(dir, "credentials"),
		SocketsDir: filepath.Join(dir, "sockets"),
		ConfigFile: filepath.Join(dir, "config"),
	}, nil
}

// SocketPath returns the socket file path for a given host name.
func (p *Paths) SocketPath(hostName string) string {
	return filepath.Join(p.SocketsDir, hostName+".sock")
}

// ExpandTilde expands ~ prefix to user home directory.
func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// EnsureDirs creates the config and sockets directories with 0700 permission.
func EnsureDirs(p *Paths) error {
	if err := os.MkdirAll(p.Dir, 0700); err != nil {
		return err
	}
	return os.MkdirAll(p.SocketsDir, 0700)
}

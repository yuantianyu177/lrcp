package cmd

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/dislab/lrcp/internal/ssh"
	"github.com/spf13/cobra"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel <from_host:port> <to_host:port>",
	Short: "Forward a port between local and remote via SSH tunnel",
	Long: `Forward a port between local and remote host via SSH tunnel.

Use localhost or 127.0.0.1 to indicate the local machine.
Any other host name is looked up from the lrcp hosts config.

Examples:
  # Forward remote server:7890 to local :9900
  lrcp tunnel server:7890 localhost:9900

  # Forward local :5432 to remote server:3243
  lrcp tunnel localhost:5432 server:3243`,
	Args: cobra.ExactArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return completeTunnelArg(toComplete)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fromHost, fromPort, err := parseTunnelEndpoint(args[0])
		if err != nil {
			return fmt.Errorf("invalid from argument: %w", err)
		}
		toHost, toPort, err := parseTunnelEndpoint(args[1])
		if err != nil {
			return fmt.Errorf("invalid to argument: %w", err)
		}

		fromLocal := isLocal(fromHost)
		toLocal := isLocal(toHost)

		if fromLocal && toLocal {
			return fmt.Errorf("both endpoints are local, nothing to tunnel")
		}
		if !fromLocal && !toLocal {
			return fmt.Errorf("both endpoints are remote, one must be local (localhost or 127.0.0.1)")
		}

		// Determine the remote host name and tunnel direction
		var remoteHostName string
		var direction ssh.TunnelDirection
		var bindAddr, targetAddr string

		if !fromLocal && toLocal {
			// Remote -> Local: SSH local forwarding (-L)
			remoteHostName = fromHost
			direction = ssh.LocalForward
			bindAddr = net.JoinHostPort(toHost, toPort)
			targetAddr = net.JoinHostPort("localhost", fromPort)
		} else {
			// Local -> Remote: SSH remote forwarding (-R)
			remoteHostName = toHost
			direction = ssh.RemoteForward
			bindAddr = net.JoinHostPort("localhost", toPort)
			targetAddr = net.JoinHostPort(fromHost, fromPort)
		}

		host, paths, err := loadHostByName(remoteHostName)
		if err != nil {
			return err
		}

		// Check sshpass if password auth
		if host.HasPassword {
			if _, err := exec.LookPath("sshpass"); err != nil {
				return fmt.Errorf("sshpass not found, install with: sudo apt install sshpass")
			}
		}

		sockPath, err := autoConnect(host, paths)
		if err != nil {
			return err
		}

		if err := ssh.Tunnel(ssh.TunnelOptions{
			SocketPath: sockPath,
			Direction:  direction,
			BindAddr:   bindAddr,
			TargetAddr: targetAddr,
		}); err != nil {
			return err
		}

		// Record the tunnel
		entry := ssh.TunnelEntry{
			Host:       host.Name,
			Direction:  direction,
			From:       args[0],
			To:         args[1],
			BindAddr:   bindAddr,
			TargetAddr: targetAddr,
			CreatedAt:  time.Now(),
		}
		if err := ssh.AddTunnel(paths.SocketsDir, entry); err != nil {
			return fmt.Errorf("save tunnel record: %w", err)
		}

		dirLabel := "remote -> local"
		if direction == ssh.RemoteForward {
			dirLabel = "local -> remote"
		}
		fmt.Printf("Tunnel started (%s): %s -> %s via %q\n", dirLabel, args[0], args[1], host.Name)
		return nil
	},
}

// parseTunnelEndpoint parses "host:port" string.
func parseTunnelEndpoint(s string) (host, port string, err error) {
	host, port, err = net.SplitHostPort(s)
	if err != nil {
		return "", "", fmt.Errorf("expected host:port, got %q", s)
	}
	if host == "" || port == "" {
		return "", "", fmt.Errorf("expected host:port, got %q", s)
	}
	return host, port, nil
}

// isLocal returns true if the host refers to the local machine.
func isLocal(host string) bool {
	return host == "localhost" || host == "127.0.0.1"
}

// completeTunnelArg provides completion for tunnel endpoints (host:port).
func completeTunnelArg(toComplete string) ([]string, cobra.ShellCompDirective) {
	// If already contains ":", no more host completion needed
	if strings.Contains(toComplete, ":") {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Offer localhost and remote host names with trailing ":"
	candidates := []string{"localhost:"}
	for _, name := range hostNames() {
		candidate := name + ":"
		if strings.HasPrefix(candidate, toComplete) {
			candidates = append(candidates, candidate)
		}
	}
	return candidates, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

func init() {
	rootCmd.AddCommand(tunnelCmd)
}

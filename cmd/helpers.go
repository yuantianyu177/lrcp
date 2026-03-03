package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/dislab/lrcp/internal/completion"
	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/crypto"
	"github.com/dislab/lrcp/internal/rsync"
	"github.com/dislab/lrcp/internal/ssh"
	"github.com/spf13/cobra"
)

// promptWithDefault displays a prompt with optional default and reads user input.
func promptWithDefault(reader *bufio.Reader, label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}
	return line
}

// runTransfer is the shared logic for push and pull commands.
func runTransfer(cmd *cobra.Command, args []string, direction rsync.Direction) error {
	var localPath, hostRemote string
	if direction == rsync.Push {
		localPath, hostRemote = args[0], args[1]
	} else {
		hostRemote, localPath = args[0], args[1]
	}

	hostName, remotePath, err := parseHostPath(hostRemote)
	if err != nil {
		return err
	}

	host, paths, err := loadHostByName(hostName)
	if err != nil {
		return err
	}
	if err := config.EnsureDirs(paths); err != nil {
		return err
	}

	sockPath, err := autoConnect(host, paths)
	if err != nil {
		return err
	}

	configArgs, err := config.ParseRsyncConfig(paths.ConfigFile)
	if err != nil {
		return err
	}

	var extraArgs []string
	if cmd.ArgsLenAtDash() >= 0 {
		extraArgs = args[cmd.ArgsLenAtDash():]
	}

	return rsync.Transfer(rsync.TransferOptions{
		Host:       host,
		LocalPath:  localPath,
		RemotePath: remotePath,
		Direction:  direction,
		ConfigArgs: configArgs,
		ExtraArgs:  extraArgs,
		SocketPath: sockPath,
	})
}

// autoConnect ensures a host is connected, returns socketPath.
func autoConnect(host *config.Host, paths *config.Paths) (string, error) {
	sockPath := paths.SocketPath(host.Name)
	if ssh.Check(sockPath) {
		return sockPath, nil
	}

	// Decrypt password if needed
	var password string
	if host.HasPassword {
		creds, err := config.LoadCredentials(paths.CredsFile)
		if err != nil {
			return "", fmt.Errorf("load credentials: %w", err)
		}
		enc, ok := creds[host.Name]
		if !ok {
			return "", fmt.Errorf("no password stored for host %q", host.Name)
		}
		key, err := crypto.DeriveKey()
		if err != nil {
			return "", err
		}
		password, err = crypto.Decrypt(enc, key)
		if err != nil {
			return "", fmt.Errorf("decrypt password: %w", err)
		}
	}

	if err := ssh.Connect(host, password, sockPath); err != nil {
		return "", err
	}
	return sockPath, nil
}

// loadHostByName loads config and finds a host by name.
func loadHostByName(name string) (*config.Host, *config.Paths, error) {
	paths, err := config.GetPaths()
	if err != nil {
		return nil, nil, err
	}
	hosts, err := config.ParseHosts(paths.HostsFile)
	if err != nil {
		return nil, nil, err
	}
	host, err := config.FindHost(hosts, name)
	if err != nil {
		return nil, nil, err
	}
	return host, paths, nil
}

// hostNames returns all host names for completion.
func hostNames() []string {
	paths, err := config.GetPaths()
	if err != nil {
		return nil
	}
	hosts, err := config.ParseHosts(paths.HostsFile)
	if err != nil {
		return nil
	}
	names := make([]string, len(hosts))
	for i, h := range hosts {
		names[i] = h.Name
	}
	return names
}

// completeHostRemote returns host:path completions for push/pull.
// Before ":", it completes host names. After ":", it completes remote paths.
func completeHostRemote(toComplete string) ([]string, cobra.ShellCompDirective) {
	// If input contains ":", complete remote path
	if idx := strings.Index(toComplete, ":"); idx >= 0 {
		hostName := toComplete[:idx]
		pathPrefix := toComplete[idx+1:]

		host, paths, err := loadHostByName(hostName)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		sockPath := paths.SocketPath(host.Name)
		candidates, directive := completion.RemotePathCompletion(host, sockPath, pathPrefix)

		// Prepend "host:" to each candidate
		var results []string
		for _, c := range candidates {
			results = append(results, hostName+":"+c)
		}
		return results, directive
	}

	// No ":" yet, complete host names with trailing ":"
	names := hostNames()
	var results []string
	for _, n := range names {
		candidate := n + ":"
		if strings.HasPrefix(candidate, toComplete) {
			results = append(results, candidate)
		}
	}
	return results, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}


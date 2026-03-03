package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/crypto"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new host configuration interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths, err := config.GetPaths()
		if err != nil {
			return err
		}
		if err := config.EnsureDirs(paths); err != nil {
			return err
		}
		if err := config.EnsureRsyncConfig(paths.ConfigFile); err != nil {
			return err
		}

		hosts, err := config.ParseHosts(paths.HostsFile)
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)
		prompt := func(label, def string) string {
			return promptWithDefault(reader, label, def)
		}

		name := prompt("Name", "")
		if name == "" {
			return fmt.Errorf("name is required")
		}
		if _, err := config.FindHost(hosts, name); err == nil {
			return fmt.Errorf("host %q already exists", name)
		}

		hostname := prompt("HostName", "")
		if hostname == "" {
			return fmt.Errorf("hostname is required")
		}
		user := prompt("User", "")
		if user == "" {
			return fmt.Errorf("user is required")
		}
		portStr := prompt("Port", "22")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return fmt.Errorf("invalid port: %s", portStr)
		}

		authType := prompt("Auth type (key/password)", "key")

		host := config.Host{
			Name:     name,
			HostName: hostname,
			User:     user,
			Port:     port,
		}

		if authType == "password" {
			fmt.Print("Password: ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}
			key, err := crypto.DeriveKey()
			if err != nil {
				return err
			}
			enc, err := crypto.Encrypt(string(pw), key)
			if err != nil {
				return err
			}
			creds, err := config.LoadCredentials(paths.CredsFile)
			if err != nil {
				return err
			}
			creds[name] = enc
			if err := config.SaveCredentials(paths.CredsFile, creds); err != nil {
				return err
			}
			host.HasPassword = true
		} else {
			keyPath, err := promptIdentityFile()
			if err != nil {
				return err
			}
			host.IdentityFile = keyPath
		}

		hosts = append(hosts, host)
		if err := config.WriteHosts(paths.HostsFile, hosts); err != nil {
			return err
		}
		fmt.Printf("Host %q added.\n", name)
		return nil
	},
}

// listSSHKeys returns private key files in ~/.ssh/ (excludes .pub, known_hosts, config, authorized_keys).
func listSSHKeys() []readline.PrefixCompleterInterface {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	sshDir := filepath.Join(home, ".ssh")
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		return nil
	}
	skip := map[string]bool{
		"known_hosts": true, "known_hosts.old": true,
		"config": true, "authorized_keys": true,
	}
	var items []readline.PrefixCompleterInterface
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".pub") || skip[e.Name()] {
			continue
		}
		items = append(items, readline.PcItem("~/.ssh/"+e.Name()))
	}
	return items
}

// expandKeyPath expands ~ prefix and validates the key file exists.
func expandKeyPath(keyPath string) (string, error) {
	expanded := config.ExpandTilde(keyPath)
	if _, err := os.Stat(expanded); err != nil {
		return "", fmt.Errorf("key file not found: %s", keyPath)
	}
	return keyPath, nil
}

// promptIdentityFile prompts for an identity file path with tab completion and retry.
func promptIdentityFile() (string, error) {
	return promptIdentityFileWithDefault("~/.ssh/id_rsa")
}

// promptIdentityFileWithDefault prompts with a custom default value.
func promptIdentityFileWithDefault(def string) (string, error) {
	if def == "" {
		def = "~/.ssh/id_rsa"
	}
	completer := readline.NewPrefixCompleter(listSSHKeys()...)
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       fmt.Sprintf("IdentityFile [%s]: ", def),
		AutoComplete: completer,
		HistoryFile:  "",
	})
	if err != nil {
		return "", err
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			return "", fmt.Errorf("read identity file: %w", err)
		}
		keyPath := strings.TrimSpace(line)
		if keyPath == "" {
			keyPath = def
		}
		result, err := expandKeyPath(keyPath)
		if err != nil {
			fmt.Printf("Error: %s, please try again.\n", err)
			continue
		}
		return result, nil
	}
}

func init() {
	rootCmd.AddCommand(newCmd)
}

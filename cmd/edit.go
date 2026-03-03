package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/crypto"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var editCmd = &cobra.Command{
	Use:   "edit <host>",
	Short: "Edit an existing host configuration",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return hostNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		paths, err := config.GetPaths()
		if err != nil {
			return err
		}
		hosts, err := config.ParseHosts(paths.HostsFile)
		if err != nil {
			return err
		}
		host, err := config.FindHost(hosts, args[0])
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)
		prompt := func(label, def string) string {
			return promptWithDefault(reader, label, def)
		}

		newName := prompt("Host", host.Name)
		if newName != host.Name {
			if _, err := config.FindHost(hosts, newName); err == nil {
				return fmt.Errorf("host %q already exists", newName)
			}
			// Update credentials key if password auth
			if host.HasPassword {
				creds, _ := config.LoadCredentials(paths.CredsFile)
				if enc, ok := creds[host.Name]; ok {
					creds[newName] = enc
					delete(creds, host.Name)
					config.SaveCredentials(paths.CredsFile, creds)
				}
			}
			host.Name = newName
		}

		host.HostName = prompt("HostName", host.HostName)
		host.User = prompt("User", host.User)
		portStr := prompt("Port", strconv.Itoa(host.Port))
		if p, err := strconv.Atoi(portStr); err == nil {
			host.Port = p
		}

		authType := "key"
		if host.HasPassword {
			authType = "password"
		}
		authType = prompt("Auth type (key/password)", authType)

		if authType == "password" {
			fmt.Print("New password (Enter to keep current): ")
			pw, err := term.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
			if err != nil {
				return err
			}
			if len(pw) > 0 {
				key, err := crypto.DeriveKey()
				if err != nil {
					return err
				}
				enc, err := crypto.Encrypt(string(pw), key)
				if err != nil {
					return err
				}
				creds, _ := config.LoadCredentials(paths.CredsFile)
				creds[host.Name] = enc
				config.SaveCredentials(paths.CredsFile, creds)
			}
			host.HasPassword = true
			host.IdentityFile = ""
		} else {
			keyPath, err := promptIdentityFileWithDefault(host.IdentityFile)
			if err != nil {
				return err
			}
			host.IdentityFile = keyPath
			host.HasPassword = false
		}

		if err := config.WriteHosts(paths.HostsFile, hosts); err != nil {
			return err
		}
		fmt.Printf("Host %q updated.\n", host.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

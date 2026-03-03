package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/ssh"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove <host>",
	Aliases: []string{"rm"},
	Short:   "Remove a host configuration",
	Args:    cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return hostNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		paths, err := config.GetPaths()
		if err != nil {
			return err
		}
		hosts, err := config.ParseHosts(paths.HostsFile)
		if err != nil {
			return err
		}
		if _, err := config.FindHost(hosts, name); err != nil {
			return err
		}

		fmt.Printf("Remove host %q? [y/N]: ", name)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(line)) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}

		// Close connection if active
		sockPath := paths.SocketPath(name)
		ssh.Close(sockPath)

		// Remove from hosts
		var filtered []config.Host
		for _, h := range hosts {
			if h.Name != name {
				filtered = append(filtered, h)
			}
		}
		if err := config.WriteHosts(paths.HostsFile, filtered); err != nil {
			return err
		}

		// Remove from credentials
		creds, _ := config.LoadCredentials(paths.CredsFile)
		delete(creds, name)
		config.SaveCredentials(paths.CredsFile, creds)

		fmt.Printf("Host %q removed.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

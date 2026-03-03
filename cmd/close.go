package cmd

import (
	"fmt"

	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/ssh"
	"github.com/spf13/cobra"
)

var closeCmd = &cobra.Command{
	Use:   "close [host]",
	Short: "Close SSH connection(s)",
	Args:  cobra.MaximumNArgs(1),
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

		if len(args) == 0 {
			// Close all
			connected, _ := ssh.ListConnected(paths.SocketsDir)
			if len(connected) == 0 {
				fmt.Println("No active connections.")
				return nil
			}
			for _, name := range connected {
				sockPath := paths.SocketPath(name)
				ssh.Close(sockPath)
				fmt.Printf("Closed %q.\n", name)
			}
			return nil
		}

		name := args[0]
		sockPath := paths.SocketPath(name)
		if !ssh.Check(sockPath) {
			fmt.Printf("Host %q is not connected.\n", name)
			return nil
		}
		ssh.Close(sockPath)
		fmt.Printf("Closed %q.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(closeCmd)
}

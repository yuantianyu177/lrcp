package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <host>",
	Short: "Establish SSH ControlMaster connection to a host",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return hostNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		host, paths, err := loadHostByName(args[0])
		if err != nil {
			return err
		}

		// Check sshpass if password auth
		if host.HasPassword {
			if _, err := exec.LookPath("sshpass"); err != nil {
				return fmt.Errorf("sshpass not found, install with: sudo apt install sshpass")
			}
		}

		if _, err := autoConnect(host, paths); err != nil {
			return err
		}
		fmt.Printf("Connected to %q.\n", host.Name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

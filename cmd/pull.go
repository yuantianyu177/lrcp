package cmd

import (
	"github.com/dislab/lrcp/internal/rsync"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull <host>:<remote> <local> [-- rsync_flags...]",
	Short: "Pull remote files to local via rsync",
	Args:  cobra.MinimumNArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completeHostRemote(toComplete)
		}
		return nil, cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTransfer(cmd, args, rsync.Pull)
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}

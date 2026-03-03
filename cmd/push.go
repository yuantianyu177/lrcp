package cmd

import (
	"fmt"
	"strings"

	"github.com/dislab/lrcp/internal/rsync"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push <local> <host>:<remote> [-- rsync_flags...]",
	Short: "Push local files to remote host via rsync",
	Args:  cobra.MinimumNArgs(2),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 1 {
			return completeHostRemote(toComplete)
		}
		return nil, cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTransfer(cmd, args, rsync.Push)
	},
}

func parseHostPath(s string) (host, path string, err error) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return "", "", fmt.Errorf("invalid format %q, expected <host>:<path>", s)
	}
	return s[:idx], s[idx+1:], nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
}

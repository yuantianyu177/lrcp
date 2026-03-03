package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/ssh"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all configured hosts",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths, err := config.GetPaths()
		if err != nil {
			return err
		}
		hosts, err := config.ParseHosts(paths.HostsFile)
		if err != nil {
			return err
		}
		if len(hosts) == 0 {
			fmt.Println("No hosts configured. Use 'lrcp new' to add one.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tHOSTNAME\tUSER\tPORT\tAUTH\tSTATUS")
		for _, h := range hosts {
			auth := "key"
			if h.HasPassword {
				auth = "password"
			}
			status := "disconnected"
			sockPath := paths.SocketPath(h.Name)
			if ssh.Check(sockPath) {
				status = "connected"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
				h.Name, h.HostName, h.User, h.Port, auth, status)
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}

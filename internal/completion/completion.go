package completion

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/ssh"
	"github.com/spf13/cobra"
)

// RemotePathCompletion returns remote path candidates for tab completion.
func RemotePathCompletion(host *config.Host, socketPath string, prefix string) ([]string, cobra.ShellCompDirective) {
	if !ssh.Check(socketPath) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Use ls -1dp to list entries with trailing / for directories
	lsCmd := fmt.Sprintf("ls -1dp %s* 2>/dev/null", prefix)
	cmd := exec.Command("ssh",
		"-o", fmt.Sprintf("ControlPath=%s", socketPath),
		"-o", "ControlMaster=no",
		fmt.Sprintf("%s@%s", host.User, host.HostName),
		lsCmd,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var results []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			results = append(results, line)
		}
	}
	return results, cobra.ShellCompDirectiveNoSpace
}

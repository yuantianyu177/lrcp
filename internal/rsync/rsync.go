package rsync

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dislab/lrcp/internal/config"
)

// Direction indicates push or pull.
type Direction int

const (
	Push Direction = iota
	Pull
)

// TransferOptions holds all parameters for an rsync transfer.
type TransferOptions struct {
	Host       *config.Host
	LocalPath  string
	RemotePath string
	Direction  Direction
	ConfigArgs []string
	ExtraArgs  []string
	SocketPath string
}

// Transfer executes an rsync command for the given options.
func Transfer(opts TransferOptions) error {
	// Build SSH command for rsync -e
	sshParts := []string{
		"ssh",
		"-o", fmt.Sprintf("ControlPath=%s", opts.SocketPath),
		"-o", "ControlMaster=no",
		"-p", fmt.Sprintf("%d", opts.Host.Port),
	}
	if opts.Host.IdentityFile != "" {
		keyPath := config.ExpandTilde(opts.Host.IdentityFile)
		sshParts = append(sshParts, "-i", keyPath)
		sshParts = append(sshParts, "-o", "PasswordAuthentication=no")
	}
	sshCmd := strings.Join(sshParts, " ")

	remote := fmt.Sprintf("%s@%s:%s", opts.Host.User, opts.Host.HostName, opts.RemotePath)

	var args []string
	args = append(args, opts.ConfigArgs...)
	args = append(args, opts.ExtraArgs...)
	args = append(args, "-e", sshCmd)

	switch opts.Direction {
	case Push:
		args = append(args, opts.LocalPath, remote)
	case Pull:
		args = append(args, remote, opts.LocalPath)
	}

	cmd := exec.Command("rsync", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// rsync exit code 20 = received SIGINT/SIGTERM/SIGHUP, treat as graceful exit
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 20 {
				fmt.Println()
				return nil
			}
		}
		return fmt.Errorf("rsync failed: %w", err)
	}
	return nil
}


package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dislab/lrcp/internal/config"
)

// Connect establishes an SSH ControlMaster connection.
func Connect(host *config.Host, password string, socketPath string) error {
	if Check(socketPath) {
		return nil // already connected
	}

	// Remove stale socket if exists
	os.Remove(socketPath)

	args := []string{
		"-nNf",
		"-o", "ControlMaster=yes",
		"-o", fmt.Sprintf("ControlPath=%s", socketPath),
		"-o", "ControlPersist=yes",
		"-o", "StrictHostKeyChecking=accept-new",
		"-p", fmt.Sprintf("%d", host.Port),
	}
	if host.IdentityFile != "" {
		keyPath := config.ExpandTilde(host.IdentityFile)
		args = append(args, "-i", keyPath)
		args = append(args, "-o", "PasswordAuthentication=no")
	}
	args = append(args, fmt.Sprintf("%s@%s", host.User, host.HostName))

	var cmd *exec.Cmd
	if password != "" {
		sshpassPath, err := exec.LookPath("sshpass")
		if err != nil {
			return fmt.Errorf("sshpass not found, install with: sudo apt install sshpass")
		}
		fullArgs := append([]string{"-e", "ssh"}, args...)
		cmd = exec.Command(sshpassPath, fullArgs...)
		cmd.Env = append(os.Environ(), "SSHPASS="+password)
	} else {
		cmd = exec.Command("ssh", args...)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh connect failed: %w", err)
	}

	// Wait for socket to appear
	for i := 0; i < 20; i++ {
		if Check(socketPath) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("ssh connected but socket not ready")
}

// Check returns true if the ControlMaster socket is active.
func Check(socketPath string) bool {
	cmd := exec.Command("ssh", "-O", "check", "-o", fmt.Sprintf("ControlPath=%s", socketPath), "_")
	return cmd.Run() == nil
}

// Close terminates the ControlMaster connection.
func Close(socketPath string) error {
	if !Check(socketPath) {
		os.Remove(socketPath) // clean up stale socket
		return nil
	}
	cmd := exec.Command("ssh", "-O", "exit", "-o", fmt.Sprintf("ControlPath=%s", socketPath), "_")
	cmd.Run() // ignore error, socket may already be gone
	os.Remove(socketPath)
	return nil
}

// CloseAll closes all connections in the sockets directory.
func CloseAll(socketsDir string) error {
	entries, err := os.ReadDir(socketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sock") {
			Close(filepath.Join(socketsDir, e.Name()))
		}
	}
	return nil
}

// ListConnected returns names of hosts with active connections.
func ListConnected(socketsDir string) ([]string, error) {
	entries, err := os.ReadDir(socketsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var connected []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sock") {
			sockPath := filepath.Join(socketsDir, e.Name())
			if Check(sockPath) {
				name := strings.TrimSuffix(e.Name(), ".sock")
				connected = append(connected, name)
			}
		}
	}
	return connected, nil
}


package rsync

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/ssh"
)

const testHostAddr = "114.212.82.241"
const testUser = "dislab"
const testPort = 22
const testKey = "~/.ssh/id_ed25519"

func setupConnection(t *testing.T) (*config.Host, string) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	host := &config.Host{
		Name:         "test-server",
		HostName:     testHostAddr,
		User:         testUser,
		Port:         testPort,
		IdentityFile: testKey,
	}
	tmp := t.TempDir()
	sock := filepath.Join(tmp, "test.sock")
	if err := ssh.Connect(host, "", sock); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ssh.Close(sock) })
	return host, sock
}

func remoteExec(t *testing.T, sock string, command string) string {
	t.Helper()
	cmd := exec.Command("ssh", "-o", fmt.Sprintf("ControlPath=%s", sock), "-o", "ControlMaster=no",
		fmt.Sprintf("%s@%s", testUser, testHostAddr), command)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("remote exec %q: %v: %s", command, err, out)
	}
	return strings.TrimSpace(string(out))
}

func randSuffix() string {
	return fmt.Sprintf("%d", rand.Intn(1000000))
}

func TestPushFile(t *testing.T) {
	host, sock := setupConnection(t)
	tmp := t.TempDir()
	localFile := filepath.Join(tmp, "test.txt")
	os.WriteFile(localFile, []byte("hello lrcp"), 0644)

	remotePath := "/tmp/lrcp_test_" + randSuffix()
	t.Cleanup(func() {
		remoteExec(t, sock, "rm -f "+remotePath)
	})

	err := Transfer(TransferOptions{
		Host: host, LocalPath: localFile, RemotePath: remotePath,
		Direction: Push, SocketPath: sock,
	})
	if err != nil {
		t.Fatal(err)
	}

	content := remoteExec(t, sock, "cat "+remotePath)
	if content != "hello lrcp" {
		t.Errorf("expected 'hello lrcp', got %q", content)
	}
}

func TestPullFile(t *testing.T) {
	host, sock := setupConnection(t)
	remotePath := "/tmp/lrcp_pull_" + randSuffix()
	remoteExec(t, sock, fmt.Sprintf("echo -n 'pull test' > %s", remotePath))
	t.Cleanup(func() {
		remoteExec(t, sock, "rm -f "+remotePath)
	})

	tmp := t.TempDir()
	localFile := filepath.Join(tmp, "pulled.txt")

	err := Transfer(TransferOptions{
		Host: host, LocalPath: localFile, RemotePath: remotePath,
		Direction: Pull, SocketPath: sock,
	})
	if err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(localFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "pull test" {
		t.Errorf("expected 'pull test', got %q", string(data))
	}
}

func TestPushDirectory(t *testing.T) {
	host, sock := setupConnection(t)
	tmp := t.TempDir()

	// Create local dir structure
	localDir := filepath.Join(tmp, "src")
	os.MkdirAll(filepath.Join(localDir, "sub"), 0755)
	os.WriteFile(filepath.Join(localDir, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(localDir, "sub", "b.txt"), []byte("bbb"), 0644)

	remoteDir := "/tmp/lrcp_dir_test_" + randSuffix()
	t.Cleanup(func() {
		remoteExec(t, sock, "rm -rf "+remoteDir)
	})

	err := Transfer(TransferOptions{
		Host: host, LocalPath: localDir + "/", RemotePath: remoteDir + "/",
		Direction:  Push,
		SocketPath: sock,
		ExtraArgs:  []string{"-a"},
	})
	if err != nil {
		t.Fatal(err)
	}

	out := remoteExec(t, sock, "cat "+remoteDir+"/a.txt")
	if out != "aaa" {
		t.Errorf("expected 'aaa', got %q", out)
	}
	out = remoteExec(t, sock, "cat "+remoteDir+"/sub/b.txt")
	if out != "bbb" {
		t.Errorf("expected 'bbb', got %q", out)
	}
}

func TestPullDirectory(t *testing.T) {
	host, sock := setupConnection(t)
	remoteDir := "/tmp/lrcp_pull_dir_" + randSuffix()
	remoteExec(t, sock, fmt.Sprintf("mkdir -p %s/sub && echo -n 'x' > %s/x.txt && echo -n 'y' > %s/sub/y.txt",
		remoteDir, remoteDir, remoteDir))
	t.Cleanup(func() {
		remoteExec(t, sock, "rm -rf "+remoteDir)
	})

	tmp := t.TempDir()
	localDir := filepath.Join(tmp, "dst")

	err := Transfer(TransferOptions{
		Host: host, LocalPath: localDir + "/", RemotePath: remoteDir + "/",
		Direction:  Pull,
		SocketPath: sock,
		ExtraArgs:  []string{"-a"},
	})
	if err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(filepath.Join(localDir, "x.txt"))
	if string(data) != "x" {
		t.Errorf("expected 'x', got %q", string(data))
	}
	data, _ = os.ReadFile(filepath.Join(localDir, "sub", "y.txt"))
	if string(data) != "y" {
		t.Errorf("expected 'y', got %q", string(data))
	}
}

func TestExtraArgsDryRun(t *testing.T) {
	host, sock := setupConnection(t)
	tmp := t.TempDir()
	localFile := filepath.Join(tmp, "dry.txt")
	os.WriteFile(localFile, []byte("dry"), 0644)

	remotePath := "/tmp/lrcp_dry_" + randSuffix()

	err := Transfer(TransferOptions{
		Host: host, LocalPath: localFile, RemotePath: remotePath,
		Direction: Push, SocketPath: sock,
		ExtraArgs: []string{"--dry-run"},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify file does NOT exist remotely
	cmd := exec.Command("ssh", "-o", fmt.Sprintf("ControlPath=%s", sock), "-o", "ControlMaster=no",
		fmt.Sprintf("%s@%s", testUser, testHostAddr), "test -f "+remotePath)
	if cmd.Run() == nil {
		t.Error("file should not exist after dry run")
	}
}

func TestExtraArgsExclude(t *testing.T) {
	host, sock := setupConnection(t)
	tmp := t.TempDir()
	localDir := filepath.Join(tmp, "src")
	os.MkdirAll(localDir, 0755)
	os.WriteFile(filepath.Join(localDir, "keep.txt"), []byte("keep"), 0644)
	os.WriteFile(filepath.Join(localDir, "skip.log"), []byte("skip"), 0644)

	remoteDir := "/tmp/lrcp_exclude_" + randSuffix()
	t.Cleanup(func() {
		remoteExec(t, sock, "rm -rf "+remoteDir)
	})

	err := Transfer(TransferOptions{
		Host: host, LocalPath: localDir + "/", RemotePath: remoteDir + "/",
		Direction: Push, SocketPath: sock,
		ExtraArgs: []string{"-a", "--exclude=*.log"},
	})
	if err != nil {
		t.Fatal(err)
	}

	out := remoteExec(t, sock, "cat "+remoteDir+"/keep.txt")
	if out != "keep" {
		t.Errorf("keep.txt should exist, got %q", out)
	}

	cmd := exec.Command("ssh", "-o", fmt.Sprintf("ControlPath=%s", sock), "-o", "ControlMaster=no",
		fmt.Sprintf("%s@%s", testUser, testHostAddr), "test -f "+remoteDir+"/skip.log")
	if cmd.Run() == nil {
		t.Error(".log file should have been excluded")
	}
}

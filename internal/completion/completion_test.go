package completion

import (
	"path/filepath"
	"testing"

	"github.com/dislab/lrcp/internal/config"
	"github.com/dislab/lrcp/internal/ssh"
	"github.com/spf13/cobra"
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
		Name: "test-server", HostName: testHostAddr,
		User: testUser, Port: testPort, IdentityFile: testKey,
	}
	tmp := t.TempDir()
	sock := filepath.Join(tmp, "test.sock")
	if err := ssh.Connect(host, "", sock); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ssh.Close(sock) })
	return host, sock
}

func TestRemotePathCompletion(t *testing.T) {
	host, sock := setupConnection(t)
	results, _ := RemotePathCompletion(host, sock, "/tmp/")
	if len(results) == 0 {
		t.Error("expected non-empty results for /tmp/")
	}
	for _, r := range results {
		if !hasPrefix(r, "/tmp/") {
			t.Errorf("result %q should start with /tmp/", r)
		}
	}
}

func TestRemotePathCompletionTilde(t *testing.T) {
	host, sock := setupConnection(t)
	results, _ := RemotePathCompletion(host, sock, "~/")
	if len(results) == 0 {
		t.Error("expected non-empty results for ~/")
	}
}

func TestRemotePathCompletionNotConnected(t *testing.T) {
	host := &config.Host{Name: "fake", HostName: "127.0.0.1", User: "nobody", Port: 22}
	results, directive := RemotePathCompletion(host, "/nonexistent/fake.sock", "/tmp/")
	if len(results) != 0 {
		t.Error("expected empty results for non-connected host")
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Error("expected ShellCompDirectiveNoFileComp")
	}
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

package ssh

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTunnelDirectionConstants(t *testing.T) {
	if LocalForward == RemoteForward {
		t.Error("LocalForward and RemoteForward should be different")
	}
}

func TestLoadTunnelsEmpty(t *testing.T) {
	tmp := t.TempDir()
	entries, err := LoadTunnels(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestSaveAndLoadTunnels(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().Truncate(time.Second)
	entries := []TunnelEntry{
		{
			Host:       "server1",
			Direction:  LocalForward,
			From:       "server1:7890",
			To:         "localhost:9900",
			BindAddr:   "localhost:9900",
			TargetAddr: "localhost:7890",
			CreatedAt:  now,
		},
		{
			Host:       "server2",
			Direction:  RemoteForward,
			From:       "localhost:5432",
			To:         "server2:3243",
			BindAddr:   "localhost:3243",
			TargetAddr: "localhost:5432",
			CreatedAt:  now,
		},
	}

	if err := SaveTunnels(tmp, entries); err != nil {
		t.Fatal(err)
	}

	// Verify file permissions
	info, err := os.Stat(filepath.Join(tmp, "tunnels.json"))
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600, got %o", perm)
	}

	loaded, err := LoadTunnels(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(loaded))
	}
	if loaded[0].Host != "server1" || loaded[0].Direction != LocalForward {
		t.Errorf("entry 0 mismatch: %+v", loaded[0])
	}
	if loaded[1].Host != "server2" || loaded[1].Direction != RemoteForward {
		t.Errorf("entry 1 mismatch: %+v", loaded[1])
	}
}

func TestAddTunnel(t *testing.T) {
	tmp := t.TempDir()
	entry := TunnelEntry{
		Host: "server1",
		From: "server1:80",
		To:   "localhost:8080",
	}

	if err := AddTunnel(tmp, entry); err != nil {
		t.Fatal(err)
	}

	entries, _ := LoadTunnels(tmp)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// Add another
	entry2 := TunnelEntry{
		Host: "server2",
		From: "localhost:3306",
		To:   "server2:3306",
	}
	if err := AddTunnel(tmp, entry2); err != nil {
		t.Fatal(err)
	}

	entries, _ = LoadTunnels(tmp)
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestRemoveTunnel(t *testing.T) {
	tmp := t.TempDir()
	entries := []TunnelEntry{
		{Host: "s1", From: "s1:80", To: "localhost:8080"},
		{Host: "s2", From: "localhost:3306", To: "s2:3306"},
	}
	SaveTunnels(tmp, entries)

	if err := RemoveTunnel(tmp, "s1:80", "localhost:8080"); err != nil {
		t.Fatal(err)
	}

	remaining, _ := LoadTunnels(tmp)
	if len(remaining) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(remaining))
	}
	if remaining[0].Host != "s2" {
		t.Errorf("wrong entry remaining: %+v", remaining[0])
	}
}

func TestRemoveTunnelNotFound(t *testing.T) {
	tmp := t.TempDir()
	entries := []TunnelEntry{
		{Host: "s1", From: "s1:80", To: "localhost:8080"},
	}
	SaveTunnels(tmp, entries)

	// Remove non-existent entry should not error
	if err := RemoveTunnel(tmp, "s1:9999", "localhost:1234"); err != nil {
		t.Fatal(err)
	}

	remaining, _ := LoadTunnels(tmp)
	if len(remaining) != 1 {
		t.Fatalf("expected 1 entry unchanged, got %d", len(remaining))
	}
}

func TestCleanTunnels(t *testing.T) {
	tmp := t.TempDir()
	// All entries refer to non-existent sockets, so all should be cleaned
	entries := []TunnelEntry{
		{Host: "dead1", From: "dead1:80", To: "localhost:80"},
		{Host: "dead2", From: "localhost:90", To: "dead2:90"},
	}
	SaveTunnels(tmp, entries)

	if err := CleanTunnels(tmp); err != nil {
		t.Fatal(err)
	}

	remaining, _ := LoadTunnels(tmp)
	if len(remaining) != 0 {
		t.Errorf("expected all entries cleaned, got %d", len(remaining))
	}
}

func TestTunnelIntegrationLocalForward(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tmp := t.TempDir()
	sock := setupConnection(t, tmp)
	defer Close(sock)

	err := Tunnel(TunnelOptions{
		SocketPath: sock,
		Direction:  LocalForward,
		BindAddr:   "localhost:19900",
		TargetAddr: "localhost:7890",
	})
	if err != nil {
		t.Errorf("local forward failed: %v", err)
	}

	// Cancel the forward
	err = CancelTunnel(sock, LocalForward, "localhost:19900", "localhost:7890")
	if err != nil {
		t.Errorf("cancel forward failed: %v", err)
	}
}

func TestTunnelIntegrationRemoteForward(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tmp := t.TempDir()
	sock := setupConnection(t, tmp)
	defer Close(sock)

	err := Tunnel(TunnelOptions{
		SocketPath: sock,
		Direction:  RemoteForward,
		BindAddr:   "localhost:13243",
		TargetAddr: "localhost:5432",
	})
	if err != nil {
		t.Errorf("remote forward failed: %v", err)
	}

	// Cancel the forward
	err = CancelTunnel(sock, RemoteForward, "localhost:13243", "localhost:5432")
	if err != nil {
		t.Errorf("cancel forward failed: %v", err)
	}
}

func TestTunnelInvalidSocket(t *testing.T) {
	err := Tunnel(TunnelOptions{
		SocketPath: "/nonexistent/random.sock",
		Direction:  LocalForward,
		BindAddr:   "localhost:19900",
		TargetAddr: "localhost:7890",
	})
	if err == nil {
		t.Error("expected error for invalid socket path")
	}
}

// setupConnection creates a real SSH connection for integration tests.
func setupConnection(t *testing.T, tmpDir string) string {
	t.Helper()
	sock := tmpDir + "/test.sock"
	host := testHostConfig("tunnel-test")
	if err := Connect(host, "", sock); err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if !Check(sock) {
		t.Fatal("expected connection to be active")
	}
	return sock
}

package ssh

import (
	"path/filepath"
	"testing"

	"github.com/dislab/lrcp/internal/config"
)

const testHostAddr = "114.212.82.241"
const testUser = "dislab"
const testPort = 22
const testKey = "~/.ssh/id_ed25519"

func testHostConfig(name string) *config.Host {
	return &config.Host{
		Name:         name,
		HostName:     testHostAddr,
		User:         testUser,
		Port:         testPort,
		IdentityFile: testKey,
	}
}

func TestConnect(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tmp := t.TempDir()
	sock := filepath.Join(tmp, "test.sock")
	host := testHostConfig("test-server")

	if err := Connect(host, "", sock); err != nil {
		t.Fatal(err)
	}
	defer Close(sock)

	if !Check(sock) {
		t.Error("expected connection to be active")
	}
}

func TestCheckNotConnected(t *testing.T) {
	if !Check("/nonexistent/random.sock") == true {
		// Check should return false for non-existent socket
	}
	if Check("/nonexistent/random.sock") {
		t.Error("expected Check to return false for non-existent socket")
	}
}

func TestClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tmp := t.TempDir()
	sock := filepath.Join(tmp, "test.sock")
	host := testHostConfig("test-server")

	Connect(host, "", sock)
	if err := Close(sock); err != nil {
		t.Fatal(err)
	}
	if Check(sock) {
		t.Error("expected connection to be closed")
	}
}

func TestCloseAll(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tmp := t.TempDir()
	sock1 := filepath.Join(tmp, "test1.sock")
	sock2 := filepath.Join(tmp, "test2.sock")
	host := testHostConfig("test-server")

	Connect(host, "", sock1)
	Connect(host, "", sock2)

	if err := CloseAll(tmp); err != nil {
		t.Fatal(err)
	}
	if Check(sock1) || Check(sock2) {
		t.Error("expected all connections to be closed")
	}
}

func TestConnectAlreadyConnected(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	tmp := t.TempDir()
	sock := filepath.Join(tmp, "test.sock")
	host := testHostConfig("test-server")

	if err := Connect(host, "", sock); err != nil {
		t.Fatal(err)
	}
	defer Close(sock)

	// Second connect should not error
	if err := Connect(host, "", sock); err != nil {
		t.Errorf("second Connect should not error: %v", err)
	}
}

func TestCloseNotConnected(t *testing.T) {
	if err := Close("/nonexistent/random.sock"); err != nil {
		t.Errorf("Close on non-existent socket should not error: %v", err)
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

const sampleHosts = `Host server1
  HostName 114.212.82.241
  User dislab
  Port 22
  IdentityFile ~/.ssh/id_rsa

Host server2
  HostName 10.0.0.5
  User root
  Port 2222
  Password yes
`

func TestParseHosts(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "hosts")
	os.WriteFile(path, []byte(sampleHosts), 0600)

	hosts, err := ParseHosts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	h1 := hosts[0]
	if h1.Name != "server1" || h1.HostName != "114.212.82.241" || h1.User != "dislab" ||
		h1.Port != 22 || h1.IdentityFile != "~/.ssh/id_rsa" || h1.HasPassword {
		t.Errorf("server1 fields mismatch: %+v", h1)
	}

	h2 := hosts[1]
	if h2.Name != "server2" || h2.HostName != "10.0.0.5" || h2.User != "root" ||
		h2.Port != 2222 || h2.IdentityFile != "" || !h2.HasPassword {
		t.Errorf("server2 fields mismatch: %+v", h2)
	}
}

func TestParseHostsEmptyFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "hosts")
	os.WriteFile(path, []byte(""), 0600)

	hosts, err := ParseHosts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}

func TestParseHostsFileNotExist(t *testing.T) {
	hosts, err := ParseHosts("/nonexistent/path/hosts")
	if err != nil {
		t.Fatal(err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}

func TestWriteAndReparse(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "hosts")

	original := []Host{
		{Name: "s1", HostName: "1.2.3.4", User: "u1", Port: 22, IdentityFile: "~/.ssh/key"},
		{Name: "s2", HostName: "5.6.7.8", User: "u2", Port: 2222, HasPassword: true},
	}
	if err := WriteHosts(path, original); err != nil {
		t.Fatal(err)
	}

	parsed, err := ParseHosts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(parsed) != len(original) {
		t.Fatalf("expected %d hosts, got %d", len(original), len(parsed))
	}
	for i := range original {
		if parsed[i] != original[i] {
			t.Errorf("host %d mismatch: got %+v, want %+v", i, parsed[i], original[i])
		}
	}
}

func TestFindHost(t *testing.T) {
	hosts := []Host{
		{Name: "server1", HostName: "1.2.3.4"},
		{Name: "server2", HostName: "5.6.7.8"},
	}
	h, err := FindHost(hosts, "server1")
	if err != nil {
		t.Fatal(err)
	}
	if h.HostName != "1.2.3.4" {
		t.Errorf("expected 1.2.3.4, got %s", h.HostName)
	}
}

func TestFindHostNotFound(t *testing.T) {
	hosts := []Host{{Name: "server1"}}
	_, err := FindHost(hosts, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent host")
	}
}

func TestLoadSaveCredentials(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "creds")

	creds := map[string]string{"server1": "encrypted_pw"}
	if err := SaveCredentials(path, creds); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected 0600 perm, got %o", perm)
	}

	loaded, err := LoadCredentials(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded["server1"] != "encrypted_pw" {
		t.Errorf("credential mismatch: got %q", loaded["server1"])
	}
}

func TestLoadCredentialsFileNotExist(t *testing.T) {
	creds, err := LoadCredentials("/nonexistent/creds")
	if err != nil {
		t.Fatal(err)
	}
	if len(creds) != 0 {
		t.Errorf("expected empty map, got %d entries", len(creds))
	}
}

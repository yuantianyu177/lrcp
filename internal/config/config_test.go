package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetPaths(t *testing.T) {
	p, err := GetPaths()
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(p.Dir) {
		t.Error("Dir should be absolute")
	}
	if filepath.Base(p.Dir) != "lrcp" {
		t.Errorf("Dir should end with lrcp, got %s", p.Dir)
	}
	if filepath.Base(p.HostsFile) != "hosts" {
		t.Error("HostsFile should end with hosts")
	}
	if filepath.Base(p.CredsFile) != "credentials" {
		t.Error("CredsFile should end with credentials")
	}
	if filepath.Base(p.SocketsDir) != "sockets" {
		t.Error("SocketsDir should end with sockets")
	}
}

func TestEnsureDirs(t *testing.T) {
	tmp := t.TempDir()
	p := &Paths{
		Dir:        filepath.Join(tmp, "lrcp"),
		HostsFile:  filepath.Join(tmp, "lrcp", "hosts"),
		CredsFile:  filepath.Join(tmp, "lrcp", "credentials"),
		SocketsDir: filepath.Join(tmp, "lrcp", "sockets"),
	}
	if err := EnsureDirs(p); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(p.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if !info.IsDir() {
		t.Error("Dir should be a directory")
	}
	if perm := info.Mode().Perm(); perm != 0700 {
		t.Errorf("Dir perm should be 0700, got %o", perm)
	}
	if _, err := os.Stat(p.SocketsDir); err != nil {
		t.Fatal("SocketsDir should exist")
	}
}

func TestEnsureDirsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	p := &Paths{
		Dir:        filepath.Join(tmp, "lrcp"),
		SocketsDir: filepath.Join(tmp, "lrcp", "sockets"),
	}
	if err := EnsureDirs(p); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDirs(p); err != nil {
		t.Fatal("second EnsureDirs should not error")
	}
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRsyncConfig(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config")
	content := `# comment line
archive-mode archive true
verbose-output verbose false
bandwidth-limit bwlimit 1000
exclude-patterns exclude .git/ *.log
`
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	args, err := ParseRsyncConfig(tmp)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"--archive", "--bwlimit=1000", "--exclude=.git/", "--exclude=*.log"}
	if len(args) != len(expected) {
		t.Fatalf("got %v, want %v", args, expected)
	}
	for i := range expected {
		if args[i] != expected[i] {
			t.Errorf("args[%d] = %q, want %q", i, args[i], expected[i])
		}
	}
}

func TestParseRsyncConfigNotExist(t *testing.T) {
	args, err := ParseRsyncConfig(filepath.Join(t.TempDir(), "nofile"))
	if err != nil {
		t.Fatal(err)
	}
	if len(args) != 0 {
		t.Fatalf("expected empty slice, got %v", args)
	}
}

func TestEnsureRsyncConfig(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config")
	if err := EnsureRsyncConfig(tmp); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != DefaultRsyncConfig {
		t.Error("config content mismatch")
	}
}

func TestEnsureRsyncConfigIdempotent(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config")
	custom := "archive-mode archive true\n"
	if err := os.WriteFile(tmp, []byte(custom), 0644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureRsyncConfig(tmp); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != custom {
		t.Error("existing config was overwritten")
	}
}

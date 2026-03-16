package cmd

import "testing"

func TestParseTunnelEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		wantHost string
		wantPort string
		wantErr  bool
	}{
		{"localhost:8080", "localhost", "8080", false},
		{"127.0.0.1:9900", "127.0.0.1", "9900", false},
		{"server:7890", "server", "7890", false},
		{"myhost:22", "myhost", "22", false},
		// error cases
		{"localhost", "", "", true},
		{"", "", "", true},
		{":8080", "", "", true},
		{"host:", "", "", true},
		{"no-port", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			host, port, err := parseTunnelEndpoint(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if host != tt.wantHost {
				t.Errorf("host: got %q, want %q", host, tt.wantHost)
			}
			if port != tt.wantPort {
				t.Errorf("port: got %q, want %q", port, tt.wantPort)
			}
		})
	}
}

func TestIsLocal(t *testing.T) {
	tests := []struct {
		host string
		want bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"server", false},
		{"myhost", false},
		{"192.168.1.1", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			if got := isLocal(tt.host); got != tt.want {
				t.Errorf("isLocal(%q) = %v, want %v", tt.host, got, tt.want)
			}
		})
	}
}

func TestTunnelDirectionDetection(t *testing.T) {
	// Test that we correctly identify remote->local vs local->remote
	tests := []struct {
		name     string
		from     string
		to       string
		wantErr  string
		wantDir  string // "local" or "remote"
	}{
		{
			name:    "remote to local",
			from:    "server:7890",
			to:      "localhost:9900",
			wantDir: "local",
		},
		{
			name:    "local to remote",
			from:    "localhost:5432",
			to:      "server:3243",
			wantDir: "remote",
		},
		{
			name:    "both local",
			from:    "localhost:8080",
			to:      "127.0.0.1:9090",
			wantErr: "both endpoints are local",
		},
		{
			name:    "both remote",
			from:    "server1:8080",
			to:      "server2:9090",
			wantErr: "both endpoints are remote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fromHost, _, err := parseTunnelEndpoint(tt.from)
			if err != nil {
				t.Fatalf("parse from: %v", err)
			}
			toHost, _, err := parseTunnelEndpoint(tt.to)
			if err != nil {
				t.Fatalf("parse to: %v", err)
			}

			fromLocal := isLocal(fromHost)
			toLocal := isLocal(toHost)

			if fromLocal && toLocal {
				if tt.wantErr == "" {
					t.Error("unexpected both-local error")
				}
				return
			}
			if !fromLocal && !toLocal {
				if tt.wantErr == "" {
					t.Error("unexpected both-remote error")
				}
				return
			}

			if tt.wantErr != "" {
				t.Errorf("expected error %q but got none", tt.wantErr)
				return
			}

			if !fromLocal && toLocal {
				if tt.wantDir != "local" {
					t.Errorf("expected local forward, got %s", tt.wantDir)
				}
			} else {
				if tt.wantDir != "remote" {
					t.Errorf("expected remote forward, got %s", tt.wantDir)
				}
			}
		})
	}
}

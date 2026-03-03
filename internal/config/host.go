package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Host represents an SSH host configuration.
type Host struct {
	Name         string
	HostName     string
	User         string
	Port         int
	IdentityFile string
	HasPassword  bool
}

// ParseHosts reads an SSH-config-like hosts file and returns a slice of Host.
func ParseHosts(path string) ([]Host, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var hosts []Host
	var current *Host
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "Host ") {
			if current != nil {
				hosts = append(hosts, *current)
			}
			current = &Host{
				Name: strings.TrimSpace(strings.TrimPrefix(line, "Host ")),
				Port: 22,
			}
			continue
		}
		if current == nil {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], strings.TrimSpace(parts[1])
		switch key {
		case "HostName":
			current.HostName = val
		case "User":
			current.User = val
		case "Port":
			if p, err := strconv.Atoi(val); err == nil {
				current.Port = p
			}
		case "IdentityFile":
			current.IdentityFile = val
		case "Password":
			current.HasPassword = val == "yes"
		}
	}
	if current != nil {
		hosts = append(hosts, *current)
	}
	return hosts, scanner.Err()
}

// WriteHosts writes hosts to an SSH-config-like file.
func WriteHosts(path string, hosts []Host) error {
	var b strings.Builder
	for i, h := range hosts {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "Host %s\n", h.Name)
		fmt.Fprintf(&b, "  HostName %s\n", h.HostName)
		fmt.Fprintf(&b, "  User %s\n", h.User)
		fmt.Fprintf(&b, "  Port %d\n", h.Port)
		if h.IdentityFile != "" {
			fmt.Fprintf(&b, "  IdentityFile %s\n", h.IdentityFile)
		}
		if h.HasPassword {
			b.WriteString("  Password yes\n")
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0600)
}

// FindHost returns a pointer to the host with the given name.
func FindHost(hosts []Host, name string) (*Host, error) {
	for i := range hosts {
		if hosts[i].Name == name {
			return &hosts[i], nil
		}
	}
	return nil, fmt.Errorf("host %q not found", name)
}

// LoadCredentials reads the encrypted credentials JSON file.
func LoadCredentials(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	var creds map[string]string
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return creds, nil
}

// SaveCredentials writes the credentials map as JSON with 0600 permission.
func SaveCredentials(path string, creds map[string]string) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

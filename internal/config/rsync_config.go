package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DefaultRsyncConfig is the default rsync configuration file content.
// Format: readable_name rsync_param value...
//   - Boolean: name param true/false  → --param (when true)
//   - Value:   name param value       → --param=value
//   - List:    name param v1 v2 ...   → --param=v1 --param=v2 ...
const DefaultRsyncConfig = `# lrcp rsync configuration
# Format: readable_name rsync_param value...
# Uncomment to enable. CLI args (after --) take highest priority.

## Boolean flags (name param true/false)
# archive-mode archive true
# verbose-output verbose true
# compression compress true
# show-progress progress true
# human-readable human-readable true
# skip-newer-files update true
# delete-extraneous delete true
# keep-partial-files partial true
# recurse-dirs recursive true
# copy-symlinks links true
# preserve-permissions perms true
# preserve-times times true
# verify-checksum checksum true
# show-item-changes itemize-changes true
# show-stats stats true
# simulate dry-run true

## Value flags (name param value)
# bandwidth-limit bwlimit 1000
# max-file-size max-size 100M
# min-file-size min-size 10K
# io-timeout timeout 30

## List flags (name param item1 item2 ...)
# exclude-patterns exclude .git/ .DS_Store __pycache__/ node_modules/ *.pyc *.swp *.log
# include-patterns include *.go *.md
`

// ParseRsyncConfig reads the config file and returns rsync args.
// Each uncommented line: readable_name rsync_param value...
// - "true"/"false" as sole value → boolean flag (--param when true)
// - single value (not bool)      → --param=value
// - multiple values              → --param=v1 --param=v2 ...
func ParseRsyncConfig(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var args []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			fmt.Fprintf(os.Stderr, "Warning: invalid config line (need at least 3 fields): %s\n", line)
			continue
		}

		// fields[0] = readable name (ignored)
		param := fields[1]
		values := fields[2:]

		if len(values) == 1 && (values[0] == "true" || values[0] == "false") {
			if values[0] == "true" {
				args = append(args, "--"+param)
			}
		} else if len(values) == 1 {
			args = append(args, fmt.Sprintf("--%s=%s", param, values[0]))
		} else {
			for _, v := range values {
				args = append(args, fmt.Sprintf("--%s=%s", param, v))
			}
		}
	}
	return args, scanner.Err()
}

// EnsureRsyncConfig creates the default config file if it does not exist.
func EnsureRsyncConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, []byte(DefaultRsyncConfig), 0644)
}

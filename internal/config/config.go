// Package config reads settings from DW_* environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dw-filedrop/internal/upload"
)

// Config holds everything send needs. No secrets in here: KeyFile is a path.
type Config struct {
	Host      string
	Port      int
	User      string
	KeyFile   string
	RemoteDir string
	Recipient string
}

// FromEnv builds a Config and names every missing variable at once.
func FromEnv() (*Config, error) {
	c := &Config{
		Host:      os.Getenv("DW_SFTP_HOST"),
		User:      os.Getenv("DW_SFTP_USER"),
		KeyFile:   os.Getenv("DW_SFTP_KEY_FILE"),
		RemoteDir: os.Getenv("DW_SFTP_DIR"),
		Recipient: os.Getenv("DW_PGP_RECIPIENT"),
		Port:      22,
	}
	if c.RemoteDir == "" {
		c.RemoteDir = "/incoming"
	}
	if p := os.Getenv("DW_SFTP_PORT"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("DW_SFTP_PORT: %q is not a number", p)
		}
		c.Port = n
	}
	var missing []string
	for _, kv := range []struct{ name, val string }{
		{"DW_SFTP_HOST", c.Host},
		{"DW_SFTP_USER", c.User},
		{"DW_SFTP_KEY_FILE", c.KeyFile},
		{"DW_PGP_RECIPIENT", c.Recipient},
	} {
		if kv.val == "" {
			missing = append(missing, kv.name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing env: %s", strings.Join(missing, ", "))
	}
	return c, nil
}

// Target converts the config for the upload package.
func (c *Config) Target() upload.Target {
	return upload.Target{
		Host:      c.Host,
		Port:      c.Port,
		User:      c.User,
		KeyFile:   c.KeyFile,
		RemoteDir: c.RemoteDir,
	}
}

// Package encrypt wraps the gpg binary. We shell out on purpose: Drovewell
// publish an RSA key and gpg is already on every box that runs this.
package encrypt

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// File encrypts in for recipient and writes <name>.pgp into outDir.
func File(in, outDir, recipient string) (string, error) {
	out := filepath.Join(outDir, filepath.Base(in)+".pgp")
	cmd := exec.Command("gpg", "--batch", "--yes", "--trust-model", "always",
		"--output", out, "--encrypt", "--recipient", recipient, in)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gpg: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

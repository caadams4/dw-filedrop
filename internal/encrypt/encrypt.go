// Package encrypt wraps the gpg binary. We shell out on purpose: Drovewell
// publish an RSA key and gpg is already on every box that runs this.
package encrypt

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// File encrypts in for recipient and returns the path of the .pgp file written next to it.
func File(in, recipient string) (string, error) {
	out := in + ".pgp"
	cmd := exec.Command("gpg", "--batch", "--yes", "--trust-model", "always",
		"--output", out, "--encrypt", "--recipient", recipient, in)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gpg: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

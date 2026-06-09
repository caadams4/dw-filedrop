// Package upload pushes a file to the Drovewell SFTP endpoint using the sftp binary.
package upload

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Target says where and as whom to upload.
type Target struct {
	Host      string
	Port      int
	User      string
	KeyFile   string
	RemoteDir string
}

// Put uploads local as <name>.part and renames it when done, so the
// bureau never picks up a half-written file.
func Put(t Target, local string) error {
	base := filepath.Base(local)
	part := t.RemoteDir + "/" + base + ".part"
	final := t.RemoteDir + "/" + base

	batch, err := os.CreateTemp("", "dw-sftp-*.batch")
	if err != nil {
		return err
	}
	defer os.Remove(batch.Name())
	fmt.Fprintf(batch, "put %q %q\nrename %q %q\n", local, part, part, final)
	if err := batch.Close(); err != nil {
		return err
	}

	args := []string{"-q", "-b", batch.Name(), "-P", strconv.Itoa(t.Port), "-i", t.KeyFile, "-o", "BatchMode=yes"}
	args = append(args, t.User+"@"+t.Host)
	out, err := exec.Command("sftp", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sftp: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

package main

import (
	"bytes"
	"testing"
)

const schemaV5 = "../../schema/payroll_v5.csv"

func runArgs(args ...string) (int, string) {
	var out, errw bytes.Buffer
	rc := run(args, &out, &errw)
	return rc, out.String() + errw.String()
}

func TestNoArgs(t *testing.T) {
	if rc, _ := runArgs(); rc != exitUsage {
		t.Fatalf("rc=%d", rc)
	}
}

func TestValidateOK(t *testing.T) {
	rc, out := runArgs("validate", "-schema", schemaV5, "../../testdata/sample_v5.csv")
	if rc != exitOK {
		t.Fatalf("rc=%d out=%s", rc, out)
	}
}

func TestValidateBad(t *testing.T) {
	rc, _ := runArgs("validate", "-schema", schemaV5, "../../testdata/bad_v5.csv")
	if rc != exitInvalid {
		t.Fatalf("rc=%d", rc)
	}
}

func TestMissingSchema(t *testing.T) {
	rc, _ := runArgs("validate", "-schema", "nope.csv", "../../testdata/sample_v5.csv")
	if rc != exitSchema {
		t.Fatalf("rc=%d", rc)
	}
}

func TestSendNeedsEnv(t *testing.T) {
	for _, k := range []string{"DW_SFTP_HOST", "DW_SFTP_USER", "DW_SFTP_KEY_FILE", "DW_PGP_RECIPIENT"} {
		t.Setenv(k, "")
	}
	rc, _ := runArgs("send", "-schema", schemaV5, "-dry-run", "../../testdata/sample_v5.csv")
	if rc != exitConfig {
		t.Fatalf("rc=%d", rc)
	}
}

func TestSendDryRun(t *testing.T) {
	t.Setenv("DW_SFTP_HOST", "sftp.drovewell.com")
	t.Setenv("DW_SFTP_USER", "dw-test")
	t.Setenv("DW_SFTP_KEY_FILE", "/dev/null")
	t.Setenv("DW_PGP_RECIPIENT", "payroll-test@example.invalid")
	rc, out := runArgs("send", "-schema", schemaV5, "-dry-run", "../../testdata/sample_v5.csv")
	if rc != exitOK {
		t.Fatalf("rc=%d out=%s", rc, out)
	}
}

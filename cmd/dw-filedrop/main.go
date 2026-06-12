// dw-filedrop: validate, encrypt and upload a payroll CSV to Drovewell.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"dw-filedrop/internal/config"
	"dw-filedrop/internal/encrypt"
	"dw-filedrop/internal/schema"
	"dw-filedrop/internal/upload"
	"dw-filedrop/internal/validate"
)

const version = "0.1.0"

const defaultSchema = "schema/payroll_v5.csv"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: dw-filedrop validate [-schema FILE] FILE.csv")
	fmt.Fprintln(w, "       dw-filedrop encrypt [-recipient R] FILE.csv")
	fmt.Fprintln(w, "       dw-filedrop send [-schema FILE] FILE.csv")
}

func run(args []string, out, errw io.Writer) int {
	if len(args) < 1 {
		usage(errw)
		return exitUsage
	}
	switch args[0] {
	case "validate":
		return cmdValidate(args[1:], out, errw)
	case "encrypt":
		return cmdEncrypt(args[1:], out, errw)
	case "send":
		return cmdSend(args[1:], out, errw)
	case "version":
		fmt.Fprintln(out, "dw-filedrop", version)
		return exitOK
	}
	usage(errw)
	return exitUsage
}

func cmdValidate(args []string, out, errw io.Writer) int {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(errw)
	schemaPath := fs.String("schema", defaultSchema, "schema file")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if fs.NArg() != 1 {
		usage(errw)
		return exitUsage
	}
	return checkFile(fs.Arg(0), *schemaPath, out)
}

func cmdEncrypt(args []string, out, errw io.Writer) int {
	fs := flag.NewFlagSet("encrypt", flag.ContinueOnError)
	fs.SetOutput(errw)
	recipient := fs.String("recipient", os.Getenv("DW_PGP_RECIPIENT"), "pgp recipient (key id or email)")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if fs.NArg() != 1 || *recipient == "" {
		usage(errw)
		return exitUsage
	}
	p, err := encrypt.File(fs.Arg(0), *recipient)
	if err != nil {
		fmt.Fprintln(errw, err)
		return exitEncrypt
	}
	fmt.Fprintln(out, p)
	return exitOK
}

func cmdSend(args []string, out, errw io.Writer) int {
	fs := flag.NewFlagSet("send", flag.ContinueOnError)
	fs.SetOutput(errw)
	schemaPath := fs.String("schema", defaultSchema, "schema file")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if fs.NArg() != 1 {
		usage(errw)
		return exitUsage
	}
	path := fs.Arg(0)
	if rc := checkFile(path, *schemaPath, out); rc != exitOK {
		return rc
	}
	cfg, err := config.FromEnv()
	if err != nil {
		fmt.Fprintln(errw, err)
		return exitConfig
	}
	enc, err := encrypt.File(path, cfg.Recipient)
	if err != nil {
		fmt.Fprintln(errw, err)
		return exitEncrypt
	}
	if err := upload.Put(cfg.Target(), enc); err != nil {
		fmt.Fprintln(errw, err)
		return exitUpload
	}
	os.Remove(enc)
	fmt.Fprintf(out, "sent %s to %s:%s\n", filepath.Base(enc), cfg.Host, cfg.RemoteDir)
	return exitOK
}

func checkFile(path, schemaPath string, out io.Writer) int {
	s, err := schema.Load(schemaPath)
	if err != nil {
		fmt.Fprintln(out, "schema:", err)
		return exitSchema
	}
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(out, err)
		return exitUsage
	}
	defer f.Close()
	probs, n, err := validate.Check(f, s)
	if err != nil {
		fmt.Fprintln(out, path+":", err)
		return exitInvalid
	}
	for i, p := range probs {
		if i == 50 {
			fmt.Fprintf(out, "... and %d more\n", len(probs)-50)
			break
		}
		fmt.Fprintln(out, p)
	}
	if len(probs) > 0 {
		fmt.Fprintf(out, "%s: %d problems in %d rows\n", path, len(probs), n)
		return exitInvalid
	}
	fmt.Fprintf(out, "%s: ok, %d rows (schema v%s)\n", path, n, s.Version)
	return exitOK
}

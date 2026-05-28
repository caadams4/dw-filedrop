// Package schema loads Drovewell payroll CSV schema files.
package schema

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Column is one field of the payroll file.
type Column struct {
	Name     string
	Type     string
	Required bool
	MaxLen   int
	Pattern  *regexp.Regexp
	Enum     []string
}

// Schema is an ordered list of columns plus the version taken from the file name.
type Schema struct {
	Version string
	Columns []Column
}

var versionRe = regexp.MustCompile(`_v([0-9]+)\.csv$`)

const header = "name,type,required,max_len,pattern"

// Load reads a schema file such as schema/payroll_v5.csv.
func Load(path string) (*Schema, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comment = '#'
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(rows) < 2 || strings.Join(rows[0], ",") != header {
		return nil, fmt.Errorf("%s: bad schema header", path)
	}
	s := &Schema{}
	if m := versionRe.FindStringSubmatch(path); m != nil {
		s.Version = m[1]
	}
	for i, row := range rows[1:] {
		c := Column{Name: row[0], Type: row[1], Required: row[2] == "yes"}
		if row[3] != "" {
			n, err := strconv.Atoi(row[3])
			if err != nil {
				return nil, fmt.Errorf("%s: row %d: bad max_len %q", path, i+1, row[3])
			}
			c.MaxLen = n
		}
		if row[4] != "" {
			re, err := regexp.Compile(row[4])
			if err != nil {
				return nil, fmt.Errorf("%s: row %d: %w", path, i+1, err)
			}
			c.Pattern = re
		}
		if strings.HasPrefix(c.Type, "enum:") {
			c.Enum = strings.Split(strings.TrimPrefix(c.Type, "enum:"), "|")
		}
		s.Columns = append(s.Columns, c)
	}
	return s, nil
}

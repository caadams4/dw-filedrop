// Package validate checks a payroll CSV against a schema.
package validate

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dw-filedrop/internal/schema"
)

var decimalRe = regexp.MustCompile(`^[0-9]+\.[0-9]{2}$`)

// Problem is one thing wrong with one field (or one row).
type Problem struct {
	Line   int
	Column string
	Msg    string
}

func (p Problem) String() string {
	if p.Column == "" {
		return fmt.Sprintf("line %d: %s", p.Line, p.Msg)
	}
	return fmt.Sprintf("line %d, %s: %s", p.Line, p.Column, p.Msg)
}

// Check reads r and returns every problem found and the number of data rows seen.
func Check(r io.Reader, s *schema.Schema) ([]Problem, int, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1

	header, err := cr.Read()
	if err == io.EOF {
		return nil, 0, errors.New("empty file")
	}
	if err != nil {
		return nil, 0, err
	}
	if len(header) != len(s.Columns) {
		return []Problem{{1, "", fmt.Sprintf("header has %d columns, schema v%s has %d", len(header), s.Version, len(s.Columns))}}, 0, nil
	}
	for i, c := range s.Columns {
		if header[i] != c.Name {
			return []Problem{{1, "", fmt.Sprintf("column %d is %q, expected %q", i+1, header[i], c.Name)}}, 0, nil
		}
	}

	var probs []Problem
	rows := 0
	for {
		row, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			var pe *csv.ParseError
			if errors.As(err, &pe) {
				probs = append(probs, Problem{pe.Line, "", pe.Err.Error()})
				continue
			}
			return probs, rows, err
		}
		rows++
		line, _ := cr.FieldPos(0)
		if len(row) != len(s.Columns) {
			probs = append(probs, Problem{line, "", fmt.Sprintf("expected %d fields, got %d", len(s.Columns), len(row))})
			continue
		}
		for i, c := range s.Columns {
			if msg := checkValue(c, row[i]); msg != "" {
				probs = append(probs, Problem{line, c.Name, msg})
			}
		}
	}
	return probs, rows, nil
}

// TODO pension_scheme: check against the list Drovewell sends (never got it)
func checkValue(c schema.Column, v string) string {
	if v == "" {
		if c.Required {
			return "required"
		}
		return ""
	}
	if c.MaxLen > 0 && len(v) > c.MaxLen {
		return fmt.Sprintf("longer than %d", c.MaxLen)
	}
	if c.Pattern != nil && !c.Pattern.MatchString(v) {
		return "does not match " + c.Pattern.String()
	}
	switch {
	case c.Type == "int":
		if _, err := strconv.Atoi(v); err != nil {
			return "not an integer"
		}
	case c.Type == "decimal":
		if !decimalRe.MatchString(v) {
			return "not a decimal with 2 places"
		}
	case c.Type == "date":
		if _, err := time.Parse("2006-01-02", v); err != nil {
			return "not a date (YYYY-MM-DD)"
		}
	case c.Type == "month":
		if _, err := time.Parse("2006-01", v); err != nil {
			return "not a month (YYYY-MM)"
		}
	case strings.HasPrefix(c.Type, "enum:"):
		ok := false
		for _, e := range c.Enum {
			if e == v {
				ok = true
			}
		}
		if !ok {
			return "not one of " + strings.Join(c.Enum, "|")
		}
	}
	return ""
}

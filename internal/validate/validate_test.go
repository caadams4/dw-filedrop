package validate

import (
	"os"
	"strings"
	"testing"

	"dw-filedrop/internal/schema"
)

func load(t *testing.T, name string) *schema.Schema {
	t.Helper()
	s, err := schema.Load("../../schema/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSampleV5(t *testing.T) {
	f, err := os.Open("../../testdata/sample_v5.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	probs, n, err := Check(f, load(t, "payroll_v5.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 0 || n != 8 {
		t.Fatalf("want 0 problems in 8 rows, got %d in %d: %v", len(probs), n, probs)
	}
}

func TestBadV5(t *testing.T) {
	f, err := os.Open("../../testdata/bad_v5.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	probs, _, err := Check(f, load(t, "payroll_v5.csv"))
	if err != nil {
		t.Fatal(err)
	}
	wantLines := []int{3, 4, 5, 6}
	if len(probs) != len(wantLines) {
		t.Fatalf("want %d problems, got %d: %v", len(wantLines), len(probs), probs)
	}
	for i, p := range probs {
		if p.Line != wantLines[i] {
			t.Errorf("problem %d on line %d, want %d (%s)", i, p.Line, wantLines[i], p)
		}
	}
}

func TestWrongHeader(t *testing.T) {
	probs, _, err := Check(strings.NewReader("a,b,c\n1,2,3\n"), load(t, "payroll_v5.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(probs) != 1 || probs[0].Line != 1 {
		t.Fatalf("want one header problem, got %v", probs)
	}
}

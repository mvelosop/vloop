package schema

import (
	"os"
	"path/filepath"
	"testing"
)

// V3's real data: every kept brief review in this repository validates as
// brief-review/v1.
func TestKeptBriefReviewsValidate(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("..", "..", ".vloop", "state", "brief-reviews", "*", "*", "brief-review.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		vs, err := Validate("brief-review/v1", b)
		if err != nil {
			t.Fatal(err)
		}
		if len(vs) != 0 {
			t.Errorf("%s: %+v", filepath.ToSlash(f), vs)
		}
	}
}

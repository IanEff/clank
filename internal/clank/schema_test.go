package clank_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/ianeff/clank/internal/clank"
)

// update regenerates the golden files instead of asserting against them:
//
//	go test ./internal/clank -run Golden -update
var update = flag.Bool("update", false, "update golden files in testdata")

// TestProposeToolSpec_SchemaMatchesGolden pins the propose tool's input schema —
// the autonomy boundary the model is held to — to a checked-in golden. Any change
// to proposeInput, its json/jsonschema tags, or the FailureClass enum that shifts
// the schema fails here, for free and offline, so the boundary moves in review
// rather than in a spendy integration run.
func TestProposeToolSpec_SchemaMatchesGolden(t *testing.T) {
	var indented bytes.Buffer
	if err := json.Indent(&indented, clank.ProposeToolSpec().InputSchema, "", "  "); err != nil {
		t.Fatalf("propose schema is not valid JSON: %v", err)
	}
	got := append(indented.Bytes(), '\n')

	golden := filepath.Join("testdata", "propose_schema.json")
	if *update {
		if err := os.WriteFile(golden, got, 0o644); err != nil {
			t.Fatalf("update golden: %v", err)
		}
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden (run with -update to create it): %v", err)
	}
	if diff := cmp.Diff(string(want), string(got)); diff != "" {
		t.Errorf("propose schema drifted from golden (-want +got):\n%s", diff)
	}
}

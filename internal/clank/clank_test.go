package clank_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ianeff/clank/internal/clank"
)

func TestGate(t *testing.T) {
	t.Parallel()
	withEvidence := clank.Decision{
		Evidence: []clank.EvidenceRef{{Summary: "503 rate 12%/min on /checkout", Ref: "loki:abc"}},
	}
	noEvidence := clank.Decision{Evidence: nil}

	cases := []struct {
		name      string
		decision  clank.Decision
		openDupes []clank.Outcome
		want      clank.Verdict
	}{
		{
			name:     "rejects a decision with no evidence",
			decision: noEvidence,
			want:     clank.Verdict{Admit: false, Status: clank.StatusInsufficientEvidence},
		},
		{
			name:      "suppresses a decision that already has an open proposal",
			decision:  withEvidence,
			openDupes: []clank.Outcome{{Status: clank.StatusProposed}},
			want:      clank.Verdict{Admit: false, Status: clank.StatusSuppressedDuplicate},
		},
		{
			name:     "admits a decision with evidence and no duplicate",
			decision: withEvidence,
			want:     clank.Verdict{Admit: true, Status: clank.StatusProposed},
		},
	}

	var gate clank.ReadinessGate
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := gate.Evaluate(tc.decision, tc.openDupes)
			if got.Admit != tc.want.Admit || got.Status != tc.want.Status {
				t.Errorf("gate returned the wrong verdict for the decision\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestProposalLog_OpenRespectsTheDedpWindow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := clank.NewMemProposalLog()

	at := time.Now()
	err := log.Record(ctx, clank.Outcome{
		Decision: clank.Decision{RunID: "r1", Fingerprint: "fp-1"},
		Status:   clank.StatusProposed,
		At:       at,
	})
	if err != nil {
		t.Fatal(err)
	}

	got, _ := log.Open(ctx, "fp-1", at.Add(-time.Hour))
	if len(got) != 1 {
		t.Errorf("a proposal recorded inside the window shouild be open: want 1, got %d", len(got))
	}

	got, _ = log.Open(ctx, "fp-1", at.Add(time.Hour))
	if len(got) != 0 {
		t.Errorf("a proposal older than `since` should not be open: want 0, got %d", len(got))
	}
}

func TestStore_PendingReturnsACheckpointedTurn(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := clank.NewMemStore()

	want := clank.Turn{RunID: "r1", Step: 0, Msgs: []clank.Message{{Role: "user", Content: "hi"}}}
	if err := store.Checkpoint(ctx, want); err != nil {
		t.Fatal(err)
	}
	pending, _ := store.Pending(ctx)
	if len(pending) != 1 || pending[0].RunID != "r1" {
		t.Errorf("a checkpointed turn should come bac as pending: want [r1], got %+v", pending)
	}
}

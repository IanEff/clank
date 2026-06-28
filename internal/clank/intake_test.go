package clank_test

import (
	"context"
	"testing"
	"time"

	"github.com/ianeff/clank/internal/clank"
)

func TestIntake_AssemblesAVersionedSAO(t *testing.T) {
	t.Parallel()
	in := clank.NewIntake(fakeTopologySource(), fakeChangeSource())
	sao, err := in.Assemble(context.Background(), sigBurnAccel())
	if err != nil {
		t.Fatal(err)
	}
	if sao.Version != 1 || len(sao.Change.Events) == 0 {
		t.Errorf("intake should assemble a v1 SAO with change events: %+v", sao)
	}
}

func sigBurnAccel() clank.SignalDetection {
	return clank.SignalDetection{
		Name:          "checkout-latency-burn-accel-001",
		Fingerprint:   "fp-checkout-latency-001",
		OriginService: "checkout",
		ServiceTier:   "tier-1",
		DetectorType:  "burn_rate_acceleration",
		Divergence:    clank.Divergence{Metric: "latency_p99", Observed: 850, Baseline: 200, Confidence: 0.9, Trajectory: "accelerating"},
		Impact: clank.Impact{
			Severity:    clank.Severity{DegradationPct: 40, Trajectory: "accelerating"},
			BlastRadius: clank.BlastRadius{AffectedPct: 60, Velocity: "fast", DownstreamConsumers: 3},
		},
		DetectedAt: time.Now(),
	}
}

type fakeTopo struct {
	snap clank.TopologySnapshot
	err  error
}

func (f fakeTopo) Topology(_ context.Context, _ clank.SignalDetection) (clank.TopologySnapshot, error) {
	return f.snap, f.err
}

type fakeChange struct {
	snap clank.ChangeSnapshot
	err  error
}

func (f fakeChange) Changes(_ context.Context, _ clank.SignalDetection) (clank.ChangeSnapshot, error) {
	return f.snap, f.err
}

func fakeTopologySource() clank.TopologySource {
	return fakeTopo{}
}

func fakeChangeSource() clank.ChangeSource {
	return fakeChange{snap: clank.ChangeSnapshot{Events: []clank.ChangeEvent{
		{ID: "deploy-7f3a", Type: "deploy", Target: "checkout", Age: 12 * time.Minute},
	}}}
}

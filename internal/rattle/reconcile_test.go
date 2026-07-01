package rattle_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/ianeff/clank/internal/rattle"
)

func TestReconcile(t *testing.T) {
	t.Parallel()
	slo := rattle.SLO{ID: "ceph-rgw-availability", Object: "ceph-rgw", Tier: "tier-1", Objective: 0.999}
	cases := map[string]struct {
		samples []rattle.Sample
		wantLen int
	}{
		"Reconcile emits one Detection for an accelerating SLO": {window(1, 2, 4, 8), 1},
		"Reconcile stays silent for a steady (non-accel) climb": {window(1, 2, 3, 4), 0}, // earns its keep
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: tc.samples})
			got, err := r.Reconcile(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != tc.wantLen {
				t.Error("wrong number of Detections emitted", cmp.Diff(tc.wantLen, len(got)))
			}
		})
	}
}

func TestReconcile_SuppressesADuplicateAcrossPasses(t *testing.T) {
	t.Parallel()
	slo := rattle.SLO{ID: "ceph-rgw-availability", Object: "ceph-rgw", Tier: "tier-1", Objective: 0.999}
	r := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: window(1, 2, 4, 8)})

	first, err := r.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := r.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 || len(second) != 0 {
		t.Errorf("same firing across two passes should emit once: got %d, then %d", len(first), len(second))
	}
}

func TestReconcile_EmitsOnCorrelationEvenWithoutAcceleration(t *testing.T) {
	t.Parallel()
	slo := rattle.SLO{ID: "ceph-rgw-availability", Object: "ceph-rgw", Tier: "tier-1"}
	r := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: window(1, 2, 3, 4)})
	r.Correlation = &rattle.CorrelationDetector{MinSignals: 2}
	r.CorrelationSource = fakeMultiSignalSource{slo.ID: multiSignal(map[string][]float64{
		"retries": {1, 2, 3, 4}, "timeouts": {1, 2, 3, 4},
	})}

	got, err := r.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Error("correlation alone should have fired Reconcile", cmp.Diff(1, len(got)))
	}
}

func TestReconcile_EmitsOnEnvelopeBreachEvenWithoutAcceleration(t *testing.T) {
	t.Parallel()
	slo := rattle.SLO{ID: "ceph-rgw-availability", Object: "ceph-rgw", Tier: "tier-1"}
	r := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: window(1, 2, 3, 4)})
	r.Envelope = &rattle.EnvelopeDetector{K: 2}
	r.BaselineSource = fakeBaselineSource{slo.ID: window(1, 1.1, 0.9, 1.0, 1.05)}

	got, err := r.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Error("envelope breach alone should have fired Reconcile", cmp.Diff(1, len(got)))
	}
}

func newTestReconciler(slos []rattle.SLO, src rattle.Source) *rattle.Reconciler {
	frozen := time.Unix(1000, 0)
	return &rattle.Reconciler{
		SLOs:     slos,
		Source:   src,
		Detector: rattle.AccelerationDetector{Threshold: 0.5},
		Debounce: rattle.NewDebouncer(10 * time.Minute),
		Now:      func() time.Time { return frozen },
	}
}

type fakeSource map[string][]rattle.Sample

func (f fakeSource) BurnSamples(_ context.Context, slo rattle.SLO) ([]rattle.Sample, error) {
	return f[slo.ID], nil
}

type fakeMultiSignalSource map[string]rattle.MultiSignalWindow

func (f fakeMultiSignalSource) MultiSignals(_ context.Context, slo rattle.SLO) (rattle.MultiSignalWindow, error) {
	return f[slo.ID], nil
}

type fakeBaselineSource map[string][]rattle.Sample

func (f fakeBaselineSource) BaselineSamples(_ context.Context, slo rattle.SLO) ([]rattle.Sample, error) {
	return f[slo.ID], nil
}

func TestReconcile_SkipsAStaleWindowWhenContractSet(t *testing.T) {
	t.Parallel()
	slo := rattle.SLO{ID: "ceph-rgw-availability", Object: "ceph-rgw", Tier: "tier-1"}
	r := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: window(1, 2, 4, 8)}) // would fire
	r.Contract = &rattle.SignalContract{FreshnessBound: time.Minute}                  // frozen clock is far newer than this

	got, err := r.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Error("a stale window under contract must not fire, even though the detector would have", cmp.Diff(0, len(got)))
	}
}

func TestReconcile_AttenuatesConfidenceInsideAnExclusionWindow(t *testing.T) {
	t.Parallel()
	slo := rattle.SLO{ID: "ceph-rgw-availability", Object: "ceph-rgw", Tier: "tier-1"}
	r := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: window(1, 2, 4, 8)})
	r.Contract = &rattle.SignalContract{
		FreshnessBound:   time.Hour,
		ConfidenceFloor:  0.1,
		ExclusionWindows: []rattle.ExclusionWindow{{Start: time.Unix(0, 0), End: time.Unix(2000, 0)}},
	}

	got, err := r.Reconcile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatal("expected the detector to still fire, just at lowered confidence")
	}
	if diff := cmp.Diff(0.5, got[0].Divergence.Confidence); diff != "" {
		t.Error("wrong attenuated confidence on the emitted Detection", diff)
	}
}

func TestReconcile_EnrichesWhenSourcesPresentZeroValueWhenNil(t *testing.T) {
	t.Parallel()
	slo := rattle.SLO{ID: "x", Object: "ceph-rgw", Tier: "tier-1", Dependencies: []rattle.Dependency{{Name: "payment-gateway"}}}

	withSources := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: window(1, 2, 4, 8)})
	withSources.TopologySource = fakeTopologySource{"payment-gateway": "degraded"}
	withSources.TrafficSource = fakeTrafficSource{{AffectedPct: 0.4}}

	got, _ := withSources.Reconcile(context.Background())
	if len(got[0].Topology.Upstream) == 0 {
		t.Error("TopologySource set but Detection.Topology.Upstream still empty — wiring didn't fire")
	}
	if got[0].Traffic.AffectedPct == 0 {
		t.Error("TrafficSource set but Detection.Traffic still zero-value")
	}

	withoutSources := newTestReconciler([]rattle.SLO{slo}, fakeSource{slo.ID: window(1, 2, 4, 8)})
	got2, _ := withoutSources.Reconcile(context.Background())
	if got2[0].Topology.Upstream != nil {
		t.Error("no TopologySource set — must reproduce zero-value Topology, same as pre-W8")
	}
}

type fakeTrafficSource []rattle.TrafficSample

func (f fakeTrafficSource) TrafficSamples(_ context.Context, _ rattle.SLO) ([]rattle.TrafficSample, error) {
	return f, nil
}

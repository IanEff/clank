package clank

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type SignalDetection struct {
	Name          string
	Fingerprint   string // DEDUPE KEY assigned by rattle
	OriginService string
	ServiceTier   string
	DetectorType  string
	Divergence    Divergence
	Topology      TopologyContext
	Traffic       TrafficContext
	Impact        Impact
	ContractRef   string
	DetectedAt    time.Time
}

type Divergence struct {
	Metric     string
	Observed   float64
	Baseline   float64
	Confidence float64
	Trajectory string
}

type Impact struct {
	Severity    Severity
	BlastRadius BlastRadius
}

type TopologyContext struct {
	Upstream  []ObservedNode
	Downstrea []ObservedNode
}

type ObservedNode struct {
	Service string
	State   string // healthy | degrade | down - via rattle
}

type TrafficContext struct {
	AffectedPct    float64
	Baseline       float64
	BaselineWindow string
}

var (
	ErrTopologySource = errors.New("intake: topology source")
	ErrChangeSoure    = errors.New("intake: change source")
)

type TopologySource interface {
	Topology(ctx context.Context, sig SignalDetection) (TopologySnapshot, error)
}

type ChangeSource interface {
	Changes(ctx context.Context, sig SignalDetection) (ChangeSnapshot, error)
}

type Intake struct {
	topo   TopologySource
	change ChangeSource
}

func NewIntake(topo TopologySource, change ChangeSource) *Intake {
	return &Intake{topo: topo, change: change}
}

func (in *Intake) Assemble(ctx context.Context, sig SignalDetection) (SAO, error) {
	topo, err := in.topo.Topology(ctx, sig)
	if err != nil {
		return SAO{}, fmt.Errorf("%w: %w", ErrTopologySource, err)
	}
	change, err := in.change.Changes(ctx, sig)
	if err != nil {
		return SAO{}, fmt.Errorf("%w: %w", ErrChangeSoure, err)
	}

	return SAO{
		Version:     1,
		AssembledAt: time.Now(),
		Signal: SignalSnapshot{
			Confidence:  sig.Divergence.Confidence,
			Metric:      sig.Divergence.Metric,
			Severity:    sig.Impact.Severity,
			BlastRadius: sig.Impact.BlastRadius,
		},
		Topology: topo,
		Change:   change,
	}, nil
}

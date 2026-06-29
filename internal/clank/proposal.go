// Package clank is the Reasoning Plane: it turns one Signal into a ranked,
// deduplicated, evidence-backed ProposalSet. It selects; it does not permit,
// detect, or touch infrastructure.
package clank

import "time"

type ProposalSet struct {
	Name             string            `json:"name,omitempty"`
	SignalRef        string            `json:"signalRef,omitempty"`
	SAOSnapshot      *SAO              `json:"saoSnapshot,omitempty"`
	FailureClass     FailureClass      `json:"failureClass,omitempty"`
	CausalScores     []CausalScore     `json:"causalScores,omitempty"`
	Hypotheses       []Hypothesis      `json:"hypotheses,omitempty"`
	Evidence         []EvidenceRef     `json:"evidence,omitempty"`
	ServiceTier      string            `json:"serviceTier,omitempty"`
	Gate             *GateResult       `json:"gate,omitempty"`
	Proposals        []Candidate       `json:"proposals,omitempty"`
	Recommended      string            `json:"recommended,omitempty"`
	RankingRationale *RankingRationale `json:"rankingRationale,omitempty"`
	Status           *ProposalStatus   `json:"status,omitempty"`
}

type RankingRationale struct {
	DominantAxis   string
	VelocityWeight string
}

type ProposalStatus struct {
	Phase        string // proposed | acknowledge | acted | superseded | closed | no_action
	SupersededBy string
	Outcome      string // success | failure | unknown | partial_non_converging
	ObservedAt   time.Time
}

type Hypothesis struct {
	Name   string  `json:"name"`
	Weight float64 `json:"weight"`
}

type EvidenceRef struct {
	Tool    string
	Query   string
	Summary string
	Ref     string
	Live    bool
}

type Candidate struct {
	ID              string           `json:"id,omitempty"`
	ContractRef     string           `json:"contractRef,omitempty"`
	Confidence      float64          `json:"confidence,omitempty"`
	PredictedImpact *PredictedImpact `json:"predictedImpact,omitempty"`
	ReversalPath    *ReversalPath    `json:"reversalPath,omitempty"`
	GovernanceLevel *GovernanceLevel `json:"governanceLevel,omitempty"`
	Rank            int              `json:"rank,omitempty"`
}

type PredictedImpact struct {
	SeverityReductionPct float64
	BlastRadiusDelta     float64
	SLOEffects           map[string]string
}

type ReversalPath struct {
	Method   string
	Watching string
	Trigger  string
}

type GovernanceLevel struct {
	Band             string
	ThresholdApplied float64
}

type FailureClass string

const (
	ClassDependencySaturation FailureClass = "dependency_saturation"
	ClassTrafficShift         FailureClass = "traffic_shift"
	ClassResourceExhaustion   FailureClass = "resource_exhaustion"
	ClassUnknown              FailureClass = "unknown"
)

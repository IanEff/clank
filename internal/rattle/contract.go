package rattle

import "time"

type SignalContract struct {
	FreshnessBound       time.Duration
	MinObservationWindow int
	ConfidenceFloor      float64
	ExclusionWindows     []ExclusionWindow
}

func (c SignalContract) Attenuated(base float64, at time.Time) float64 {
	attenuated := base
	for _, ex := range c.ExclusionWindows {
		if !at.Before(ex.Start) && at.Before(ex.End) {
			attenuated = base / 2
			break
		}
	}
	if attenuated < c.ConfidenceFloor {
		return c.ConfidenceFloor
	}
	return attenuated
}

func (c SignalContract) Fresh(window []Sample, now time.Time) bool {
	if len(window) == 0 {
		return false
	}
	newest := window[len(window)-1].T
	return now.Sub(newest) <= c.FreshnessBound
}

type ExclusionWindow struct {
	Start, End time.Time
	Reason     string
}

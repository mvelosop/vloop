package driver

import (
	"fmt"
	"time"
)

// costThresholds are the shares of the ceiling at which the run warns, once each.
var costThresholds = []int{50, 80}

// sessionLine is the line a finished session prints: its phase and task, how
// long it took, what it cost and what the run has spent against the ceiling.
func sessionLine(s Spec, wall time.Duration, cost, spent, ceiling float64) string {
	who := s.Phase
	if s.Phase == PhaseGateReview {
		who = "gate review"
	}
	if s.Phase == PhaseSecondReview {
		who = "second review"
	}
	if s.Phase == PhaseWork || s.Phase == PhaseReview || s.Phase == PhaseSecondReview {
		who += " " + s.Arg
	}
	secs := int(wall.Round(time.Second) / time.Second)
	return fmt.Sprintf("   %s · %dm%02ds · $%.2f · run $%.2f / $%.2f", who, secs/60, secs%60, cost, spent, ceiling)
}

// costWarnings are the warnings for the thresholds the run's spend crossed
// going from before to after. Each threshold is crossed once, so it is warned
// once, in whichever session crossed it.
func costWarnings(before, after, ceiling float64) []string {
	var out []string
	if ceiling <= 0 {
		return nil
	}
	for _, pct := range costThresholds {
		at := ceiling * float64(pct) / 100
		if before < at && at <= after {
			out = append(out, fmt.Sprintf("   cost: $%.2f spent, %d%% of the $%.2f ceiling", after, pct, ceiling))
		}
	}
	return out
}

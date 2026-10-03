package driver

// runGateRetry runs a gate and, if it fails, once more at once with nothing
// changed: a pass on the re-run is a flaky gate, recorded and not charged.
func (it *Iterator) runGateRetry(iter int, active, id string) (*gateResult, error) {
	g, err := it.runGate(iter, active, id)
	if err != nil || g.exit == 0 {
		return g, err
	}
	g2, err := it.runGate(iter, active, id)
	if err != nil {
		return nil, err
	}
	if g2.exit == 0 {
		g2.flaky = true
		it.warn("   FLAKY GATE %s — failed, then passed on the immediate re-run", id)
	}
	return g2, nil
}

func gatePassesLine(id string) string {
	return id + " GATE PASSES while the session reports blocked — the work satisfies its own gate; the block is about something else."
}

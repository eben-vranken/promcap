package promcap

const defaultMaxSeries = 1000

type CapOpts struct {
	MaxSeries int
	// Allow restricts a label to the listed values; non-listed values overflow immediately.
	// Allowed values still consume the MaxSeries budget.
	Allow map[string][]string

	// Evict, when true, evicts the least-recently-used series to admit a new one
	// once MaxSeries is reached, instead of collapsing into the overflow series.
	// Evicted series are deleted from the metric; for counters this discards their
	// accumulated value.
	Evict bool
}

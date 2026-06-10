package promcap

const defaultMaxSeries = 1000

type CapOpts struct {
	MaxSeries int
	// Allow restricts a label to the listed values; non-listed values overflow immediately.
	// Allowed values still consume the MaxSeries budget.
	Allow map[string][]string
}

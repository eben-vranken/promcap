package promcap

const defaultMaxSeries = 1000

type CapOpts struct {
	MaxSeries int
	Allow     map[string][]string
}

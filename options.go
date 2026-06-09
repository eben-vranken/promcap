package promcap

type CapOpts struct {
	MaxSeries int
	Allow     map[string][]string
}

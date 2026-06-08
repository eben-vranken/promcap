package promcap

import "github.com/prometheus/client_golang/prometheus"

type Cap struct {
	reg prometheus.Registerer
}

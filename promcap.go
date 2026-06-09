package promcap

import (
	"github.com/prometheus/client_golang/prometheus"
)

const overflowValue = "__overflow__"

type Cap struct {
	reg prometheus.Registerer
}

func Wrap(reg prometheus.Registerer) *Cap {
	return &Cap{reg: reg}
}

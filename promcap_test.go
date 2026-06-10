package promcap

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestWrapTwiceReusesMeta(t *testing.T) {
	reg := prometheus.NewRegistry()

	c1 := Wrap(reg)
	c2 := Wrap(reg)

	if c1.cappedTotal != c2.cappedTotal {
		t.Errorf("Wrap did not reuse the meta-counter: %p != %p", c1.cappedTotal, c2.cappedTotal)
	}
}

func TestWrapPanicsOnIncompatibleRegistration(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("expected panic on incompatible registration")
		}
	}()

	reg := prometheus.NewRegistry()

	reg.MustRegister(prometheus.NewCounterVec(prometheus.CounterOpts{Name: "promcap_series_capped_total"}, []string{"different_label"}))

	Wrap(reg)
}

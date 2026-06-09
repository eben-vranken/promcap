package promcap

import (
	"strconv"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestCounterVecConcurrentAccess(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewCounterVec(prometheus.CounterOpts{Name: "request_total"}, []string{"user"}, 100)
	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv.WithLabelValues(strconv.Itoa(i)).Inc()
		}()
	}

	wg.Wait()
}

func TestLimiterResolve(t *testing.T) {
	lim := newLimiter(2)

	gotA := lim.resolve([]string{"a"})
	lim.resolve([]string{"a"})
	lim.resolve([]string{"b"})
	gotC := lim.resolve([]string{"c"})

	if gotA[0] != "a" {
		t.Errorf("value within cap: got %q, want %q", gotA[0], "a")
	}

	if gotC[0] != overflowValue {
		t.Errorf("c with full cap: got %q, want %q", gotC[0], overflowValue)
	}
}

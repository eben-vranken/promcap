package promcap

import (
	"strconv"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestCounterVecConcurrentAccess(t *testing.T) {
	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewCounterVec(prometheus.CounterOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: 100})
	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv.WithLabelValues(strconv.Itoa(i)).Inc()
		}()
	}

	wg.Wait()

	count, err := testutil.GatherAndCount(reg, "request_total")

	if err != nil {
		t.Fatalf("Fatal error: %v", err)
	}

	if count != 101 {
		t.Errorf("Gather and count got %d, want %d", count, 101)
	}

	if testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total")) != 900 {
		t.Errorf("Overflow values got %v, want %v", testutil.ToFloat64(cv.WithLabelValues("request_total")), 900)
	}
}

func TestLimiterResolve(t *testing.T) {
	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"user"}, CapOpts{MaxSeries: 2}, meta)

	gotA := lim.resolve([]string{"a"})
	lim.resolve([]string{"a"})
	lim.resolve([]string{"b"})
	gotC := lim.resolve([]string{"c"})

	if testutil.ToFloat64(meta.WithLabelValues("test_metric")) != 1 {
		t.Errorf("series capped total: got %f, want %d", testutil.ToFloat64(meta.WithLabelValues("test_metric")), 1)
	}

	if gotA[0] != "a" {
		t.Errorf("value within cap: got %q, want %q", gotA[0], "a")
	}

	if gotC[0] != overflowValue {
		t.Errorf("c with full cap: got %q, want %q", gotC[0], overflowValue)
	}
}

func TestLimiterResolveAllowList(t *testing.T) {
	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"method"}, CapOpts{MaxSeries: 100, Allow: map[string][]string{"method": {"GET", "POST"}}}, meta)

	gotGet := lim.resolve([]string{"GET"})
	gotDelete := lim.resolve([]string{"DELETE"})

	if gotGet[0] != "GET" {
		t.Errorf("Get did not resolve: got %q, want %q", gotGet[0], "GET")
	}

	if gotDelete[0] != overflowValue {
		t.Errorf("Delete did not resolve: got %q, want %q", gotDelete[0], overflowValue)
	}
}

func TestLimiterRejectsUnknownAllowList(t *testing.T) {
	defer func() {
		r := recover()

		if r == nil {
			t.Errorf("expected panic for unknown Allow key, got none")
		}
	}()

	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	_ = newLimiter("test_metric", []string{"method"}, CapOpts{MaxSeries: 100, Allow: map[string][]string{"typo": {"GET", "POST"}}}, meta)
}

func TestNewLimiterWithNoMaxSeries(t *testing.T) {
	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"method"}, CapOpts{}, meta)

	if lim.maxSeries != defaultMaxSeries {
		t.Errorf("omitted max series was not set to default value, got %d, expected %d", lim.maxSeries, defaultMaxSeries)
	}
}

func TestLimiterAllowConsumesBudget(t *testing.T) {
	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"method"}, CapOpts{MaxSeries: 1, Allow: map[string][]string{"method": {"GET", "POST"}}}, meta)

	gotGet := lim.resolve([]string{"GET"})
	gotPost := lim.resolve([]string{"POST"})

	if gotGet[0] != "GET" {
		t.Errorf("Get did not resolve: got %q, want %q", gotGet[0], "GET")
	}

	if gotPost[0] != overflowValue {
		t.Errorf("Post did not resolve: got %q, want %q", gotPost[0], overflowValue)
	}
}

func TestLimiterMultiLabelPreservesAllowed(t *testing.T) {
	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"method", "user"}, CapOpts{MaxSeries: 1, Allow: map[string][]string{"method": {"GET", "POST"}}}, meta)

	gotGet := lim.resolve([]string{"GET", "alice"})
	gotPost := lim.resolve([]string{"POST", "bob"})
	gotDelete := lim.resolve([]string{"DELETE", "carol"})

	if gotGet[0] != "GET" {
		t.Errorf("Get did not resolve: got %q, want %q", gotGet[0], "GET")
	}

	if gotPost[0] != "POST" {
		t.Errorf("Post did not resolve: got %q, want %q", gotPost[0], "POST")
	}

	if gotPost[1] != overflowValue {
		t.Errorf("Post did not resolve: got %q, want %q", gotPost[1], overflowValue)
	}

	if gotDelete[0] != overflowValue {
		t.Errorf("Delete did not resolve: got %q, want %q", gotPost[0], overflowValue)
	}
}

func TestLimiterOrderPanicsOnMissingLabel(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("expected panic for missing label, got none")
		}
	}()

	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"user"}, CapOpts{MaxSeries: 1}, meta)
	lim.order(prometheus.Labels{})
}

func TestLimiterEvictsLeastRecentlyUsed(t *testing.T) {
	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"user"}, CapOpts{MaxSeries: 2, Evict: true}, meta)

	lim.resolve([]string{"a"})
	lim.resolve([]string{"b"})

	lim.resolve([]string{"a"})

	lim.resolve([]string{"c"})

	if _, ok := lim.seen["a"]; !ok {
		t.Errorf("recently-used series a was evicted")
	}

	if _, ok := lim.seen["c"]; !ok {
		t.Errorf("newly admitted series c is missing")
	}

	if _, ok := lim.seen["b"]; ok {
		t.Errorf("least-recently-used series b was not evicted")
	}
}

func TestLimiterAllowedLabelSurvivesUserOverflow(t *testing.T) {
	meta := prometheus.NewCounterVec(prometheus.CounterOpts{Name: "test_metric"}, []string{"metric"})
	lim := newLimiter("test_metric", []string{"method", "user"}, CapOpts{MaxSeries: 1, Allow: map[string][]string{"method": {"GET"}}}, meta)

	lim.resolve([]string{"GET", "alice"})
	got := lim.resolve([]string{"GET", "bob"})

	if got[0] != "GET" {
		t.Errorf("Get did not resolve: got %q, want %q", got[0], "GET")
	}

	if got[1] != overflowValue {
		t.Errorf("user did not collapse: got %q, want %q", got[1], overflowValue)
	}

	if testutil.ToFloat64(meta.WithLabelValues("test_metric")) != 1 {
		t.Errorf("series capped total: got %f, want %d", testutil.ToFloat64(meta.WithLabelValues("test_metric")), 1)
	}
}

func TestCounterVecParallelInvariants(t *testing.T) {
	maxSeries := 100
	const observations = 1000

	reg := prometheus.NewRegistry()
	regWrap := Wrap(reg)

	cv := regWrap.NewCounterVec(prometheus.CounterOpts{Name: "request_total"}, []string{"user"}, CapOpts{MaxSeries: maxSeries})
	wg := sync.WaitGroup{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cv.WithLabelValues(strconv.Itoa(i)).Inc()
		}()
	}

	wg.Wait()

	gathered, _ := testutil.GatherAndCount(reg, "request_total")
	capped := testutil.ToFloat64(regWrap.cappedTotal.WithLabelValues("request_total"))
	admitted := gathered - 1

	if gathered > maxSeries+1 {
		t.Errorf("Gather and count got %d, want %d", gathered, maxSeries+1)
	}

	if capped != float64(observations-admitted) {
		t.Errorf("Gather and count got %f, want %d", capped, observations-admitted)
	}
}

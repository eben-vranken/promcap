<h1 align="center">Promcap</h1>

<p align="center">
  Drop-in Prometheus <code>*Vec</code> wrapper that caps metric cardinality at the source, before unbounded labels OOM your monitoring stack.
</p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/eben-vranken/promcap"><img src="https://goreportcard.com/badge/github.com/eben-vranken/promcap" alt="Go Report Card"></a>
  <a href="https://codecov.io/gh/eben-vranken/promcap"><img src="https://codecov.io/gh/eben-vranken/promcap/branch/main/graph/badge.svg" alt="Coverage"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="MIT License"></a>
</p>

Promcap wraps Prometheus `CounterVec`, `GaugeVec`, `HistogramVec`, and
`SummaryVec` with a hard cardinality cap. Once a metric has emitted its
configured number of distinct label combinations, every further combination
collapses into a single `__overflow__` series instead of creating a new one.
Your dashboards keep working, and a runaway label (a user ID, a request path, an
attacker-controlled header) can no longer grow your time-series count without
bound.

A high-cardinality label is the classic way to take down a Prometheus stack: one
mislabelled metric quietly spawns hundreds of thousands of series until the
scrape target, the TSDB, or both run out of memory. The usual fixes are
after-the-fact (relabel rules, recording-rule drops, alerts on series
growth), and they fire once the damage is already in flight. Promcap enforces
the ceiling in-process, at the moment the series would be created, so the
unbounded growth never reaches the registry.

## Install

```sh
go get github.com/eben-vranken/promcap
```

## Quick start

Wrap a `prometheus.Registerer` once, then create capped metrics from it exactly
as you would with the upstream constructors, plus a `CapOpts`:

```go
package main

import (
	"net/http"

	"github.com/eben-vranken/promcap"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	reg := prometheus.NewRegistry()
	cap := promcap.Wrap(reg)

	requests := cap.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by route and status.",
		},
		[]string{"route", "status"},
		promcap.CapOpts{MaxSeries: 1000},
	)

	// Use it like any *CounterVec.
	requests.WithLabelValues("/checkout", "200").Inc()
	requests.With(prometheus.Labels{"route": "/checkout", "status": "500"}).Inc()

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	http.ListenAndServe(":8080", nil)
}
```

Once 1000 distinct `(route, status)` pairs have been seen, the 1001st and every
new pair after it are recorded under `route="__overflow__"`,
`status="__overflow__"` instead of minting fresh series. The collapsed
observations are still counted, just bucketed together.

## How it works

Each capped metric carries a small limiter that tracks the distinct label
combinations it has admitted:

1. The combination is checked against any per-label `Allow` lists. A value that
   is not on its label's allowlist overflows immediately, before it can consume
   the budget.
2. If the combination has been seen before, it passes straight through to the
   underlying metric.
3. If it is new and the metric is **below** `MaxSeries`, it is admitted and
   remembered.
4. If it is new and the metric is **at** `MaxSeries`, it collapses into the
   `__overflow__` series (or, with `Evict`, displaces the least-recently-used
   series; see below).

Every collapsed observation increments `promcap_series_capped_total`, a counter
labelled by `metric` that Promcap registers once per registry. Scrape it to see
exactly which metric is shedding cardinality and how much:

```promql
rate(promcap_series_capped_total[5m])
```

The limiter is guarded by a mutex, so all capped methods are safe for
concurrent use. The hot path, a label combination that has already been
admitted, takes a lock, hits a map, and returns with zero allocations.

> **Reserved value:** `__overflow__` is reserved. A real label value equal to
> `__overflow__` is indistinguishable from the overflow bucket and will merge
> into it.

## Drop-in scope

Promcap wraps the mutating and lookup methods that create series:

**Capped:** `WithLabelValues`, `With`, `GetMetricWith`,
`GetMetricWithLabelValues`, `Reset`.

**Not yet wrapped:** `CurryWith`, `Delete`, `DeleteLabelValues`. Code that
depends on these is not yet a drop-in replacement.

The capped types implement `prometheus.Collector`, so you register them on the
wrapped registry (Promcap does this for you in the `New*Vec` constructors) and
scrape them like any other collector.

## Options

```go
promcap.CapOpts{
	// MaxSeries is the cap on distinct admitted label combinations.
	// Defaults to 1000 when zero or negative.
	MaxSeries: 1000,

	// Allow restricts a label to a fixed set of values; any value not listed
	// overflows immediately. Allowed values still consume the MaxSeries budget.
	Allow: map[string][]string{
		"status": {"200", "400", "404", "500"},
	},

	// Evict, when true, evicts the least-recently-used series to make room for
	// a new one once MaxSeries is reached, instead of collapsing into the
	// overflow series. Evicted series are deleted from the metric; for counters
	// this discards their accumulated value.
	Evict: false,
}
```

### Allow lists

Use `Allow` for labels whose valid values you know up front (HTTP status codes,
a closed set of regions, a handful of event types). Anything outside the list
overflows the instant it appears, so a typo or an injected value can never even
start filling the budget:

```go
cap.NewCounterVec(
	prometheus.CounterOpts{Name: "events_total"},
	[]string{"region", "kind"},
	promcap.CapOpts{
		MaxSeries: 500,
		Allow: map[string][]string{
			"region": {"us-east", "us-west", "eu-central"},
		},
	},
)
```

A value passed for an `Allow` label that is not one of the metric's labels
panics at construction time: it is a programming error, not a runtime
condition.

### Eviction vs. overflow

By default, reaching `MaxSeries` is permanent for the run: new combinations
collapse into `__overflow__` and the admitted set never changes until `Reset`.
That is the safe choice for unbounded or adversarial labels.

Set `Evict: true` when the live set of interesting label values rotates over
time (active tenants, recently-seen hosts) and you would rather track the *most
recent* `MaxSeries` of them than freeze the first ones you happened to see.
Admission then evicts the least-recently-used series (using a clock
second-chance policy so a still-active series gets one reprieve before it is
dropped) and **deletes it from the metric**. For a counter, the evicted
series' accumulated total is discarded.

## Benchmarks

`go test -bench . -benchmem` on an AMD Ryzen 5 5600X (Go 1.26):

| Path | ns/op | B/op | allocs/op |
|------|------:|-----:|----------:|
| Admitted combination (hot path) | ~14 | 0 | 0 |
| Overflow (cap reached) | ~82 | 23 | 1 |
| New admission (under cap) | ~498 | 196 | 3 |
| Eviction flood (`Evict: true`) | ~234 | 112 | 3 |
| Mixed read/write, parallel (12 cores) | ~32 | 0 | 0 |

The case that matters in steady state, a label combination that has already
been admitted, resolves in about **14 ns with zero allocations**, so the cap
adds essentially nothing to a metric that is behaving. Because that hot path
takes only a read lock, it scales across cores instead of serializing: the
mixed read/write parallel workload resolves in **~32 ns/op on 12 cores, down
from ~347 ns** when every call contended on a single mutex. The expensive
paths are the ones you want to be rare: minting a brand-new series, or
churning the working set under eviction.

Reproduce with:

```sh
go test -bench . -benchmem -run '^$'
```

## Testing

```sh
go test ./...
```

The suite covers the limiter, every capped `*Vec` type, the allow/overflow and
eviction interactions, and concurrent access, and runs under the race detector
in CI.

## License

MIT. See [LICENSE](LICENSE).

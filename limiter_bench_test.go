package promcap

import (
	"strconv"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func BenchmarkResolveHit(b *testing.B) {
	reg := prometheus.NewRegistry()
	cv := Wrap(reg).NewCounterVec(
		prometheus.CounterOpts{Name: "request_total"},
		[]string{"user"},
		CapOpts{MaxSeries: 16},
	)
	cv.lim.resolve([]string{"hot"})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv.lim.resolve([]string{"hot"})
	}
}

func BenchmarkResolveAdmit(b *testing.B) {
	reg := prometheus.NewRegistry()
	cv := Wrap(reg).NewCounterVec(
		prometheus.CounterOpts{Name: "request_total"},
		[]string{"user"},
		CapOpts{MaxSeries: b.N + 1},
	)
	keys := make([][]string, b.N)
	for i := range keys {
		keys[i] = []string{strconv.Itoa(i)}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv.lim.resolve(keys[i])
	}
}

func BenchmarkResolveOverflow(b *testing.B) {
	reg := prometheus.NewRegistry()
	cv := Wrap(reg).NewCounterVec(
		prometheus.CounterOpts{Name: "request_total"},
		[]string{"user"},
		CapOpts{MaxSeries: 1},
	)
	cv.lim.resolve([]string{"seed"})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv.lim.resolve([]string{strconv.Itoa(i)})
	}
}

func BenchmarkResolveEvictFlood(b *testing.B) {
	reg := prometheus.NewRegistry()
	cv := Wrap(reg).NewCounterVec(
		prometheus.CounterOpts{Name: "request_total"},
		[]string{"user"},
		CapOpts{MaxSeries: 128, Evict: true},
	)
	keys := make([][]string, b.N)
	for i := range keys {
		keys[i] = []string{strconv.Itoa(i)}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv.lim.resolve(keys[i])
	}
}

func BenchmarkResolveParallel(b *testing.B) {
	reg := prometheus.NewRegistry()
	cv := Wrap(reg).NewCounterVec(
		prometheus.CounterOpts{Name: "request_total"},
		[]string{"user"},
		CapOpts{MaxSeries: 1 << 20},
	)

	const pool = 1024
	keys := make([][]string, pool)
	for i := range keys {
		keys[i] = []string{strconv.Itoa(i)}
		cv.lim.resolve(keys[i])
	}

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cv.lim.resolve(keys[i&(pool-1)])
			i++
		}
	})
}

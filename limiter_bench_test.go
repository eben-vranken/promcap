package promcap

import (
	"strconv"
	"sync/atomic"
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

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv.lim.resolve([]string{strconv.Itoa(i)})
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

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cv.lim.resolve([]string{strconv.Itoa(i)})
	}
}

func BenchmarkResolveParallel(b *testing.B) {
	reg := prometheus.NewRegistry()
	cv := Wrap(reg).NewCounterVec(
		prometheus.CounterOpts{Name: "request_total"},
		[]string{"user"},
		CapOpts{MaxSeries: 1 << 20},
	)
	cv.lim.resolve([]string{"hot"})

	var gid int64
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		id := atomic.AddInt64(&gid, 1)
		prefix := strconv.FormatInt(id, 10) + "-"
		i := 0
		for pb.Next() {
			if i&1 == 0 {
				cv.lim.resolve([]string{"hot"})
			} else {
				cv.lim.resolve([]string{prefix + strconv.Itoa(i)})
			}
			i++
		}
	})
}

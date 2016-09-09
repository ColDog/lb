package stats

import (
	"time"
	"sync"
)

type StatsCollectorBackend string

const (
	MEMORY StatsCollectorBackend = "memory"
	NOOP StatsCollectorBackend = "noop"
)

type StatsCollector interface {
	SetIncrement(key string, amount int)
	SetTime(key string, t time.Time)
	SetPoint(key string, value float64)
	GetIncrement(key string) int64
}

type NoOpStatsCollector struct {}

func (n *NoOpStatsCollector) SetIncrement(key string, amount int) {}
func (n *NoOpStatsCollector) SetTime(key string, t time.Time) {}
func (n *NoOpStatsCollector) SetPoint(key string, value float64) {}
func (n *NoOpStatsCollector) GetIncrement(key string) int64 { return int64(0)}

type point struct {
	t time.Time
	p float64
}

func New(t StatsCollectorBackend) StatsCollector {
	if t == MEMORY {
		return &InMemStatsCollector{
			lockIncr: &sync.RWMutex{},
			increments: map[string]int64{},
			lockPts: &sync.Mutex{},
			points: map[string][]point{},
			lockTms: &sync.Mutex{},
			timer: map[string][]time.Duration{},
		}
	} else {
		return &NoOpStatsCollector{}
	}
}

type InMemStatsCollector struct {
	lockIncr *sync.RWMutex
	increments map[string]int64

	lockPts *sync.Mutex
	points map[string][]point

	lockTms *sync.Mutex
	timer map[string][]time.Duration
}

func (l *InMemStatsCollector) SetIncrement(key string, amount int) {
	l.lockIncr.Lock()
	defer l.lockIncr.Unlock()
	l.increments[key] += int64(amount)
}

func (l *InMemStatsCollector) SetPoint(key string, value float64) {
	l.lockPts.Lock()
	defer l.lockPts.Unlock()
	l.points[key] = append(l.points[key], point{time.Now(), value})
}

func (l *InMemStatsCollector) SetTime(key string, t time.Time) {
	l.lockTms.Lock()
	defer l.lockTms.Unlock()
	l.timer[key] = append(l.timer[key], time.Duration(time.Now().UnixNano() - t.UnixNano()))
}

func (l *InMemStatsCollector) GetIncrement(key string) int64 {
	l.lockIncr.RLock()
	defer l.lockIncr.RUnlock()
	return l.increments[key]
}

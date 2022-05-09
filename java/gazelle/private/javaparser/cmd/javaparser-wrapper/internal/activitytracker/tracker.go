package activitytracker

import (
	"sync"
	"time"
)

type Tracker struct {
	lastActivityLock sync.RWMutex
	lastActivity     time.Time
}

func NewTracker() *Tracker {
	return &Tracker{
		lastActivity: time.Now(),
	}
}

func (t *Tracker) Ping() {
	t.lastActivityLock.Lock()
	defer t.lastActivityLock.Unlock()

	t.lastActivity = time.Now()
}

func (t *Tracker) IdleSince(q time.Time) bool {
	t.lastActivityLock.RLock()
	defer t.lastActivityLock.RUnlock()

	return t.lastActivity.Before(q)
}

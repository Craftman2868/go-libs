package clock

import "time"

const TPS_UPDATE_INTERVAL = time.Second / 4

type Clock struct {
	TickTime time.Duration
	lastTick time.Time
	// lastTickTime time.Duration
	TickCount uint64

	countStart          time.Time
	tickCountSinceStart uint32
	curTps              float32
}

func NewClock(tps int) Clock {
	var clock Clock

	clock.SetTps(tps)
	clock.TickCount = 0

	return clock
}

func (clock *Clock) UpdateTps() {
	clock.curTps = float32(clock.tickCountSinceStart) * float32(time.Second) / float32(time.Since(clock.countStart))
	clock.countStart = time.Now()
	clock.tickCountSinceStart = 0
}

func (clock *Clock) Tick() bool {
	if clock.TickTime != 0 {
		if d := time.Since(clock.lastTick); d < clock.TickTime {
			return false
		}
	}
	clock.tick()
	return true
}

func (clock *Clock) TickSleep() {
	if clock.TickTime != 0 {
		if d := time.Since(clock.lastTick); d < clock.TickTime {
			time.Sleep(clock.TickTime - d)
		}
	}
	clock.tick()
}

func (clock *Clock) TickGetTime() time.Duration {
	if clock.TickTime != 0 {
		if d := time.Since(clock.lastTick); d < clock.TickTime {
			return clock.TickTime - d
		}
	}
	clock.tick()
	return 0
}

func (clock *Clock) tick() {
	// clock.lastTickTime = time.Since(clock.lastTick)
	clock.lastTick = time.Now()
	clock.TickCount++
	clock.tickCountSinceStart++
	if time.Since(clock.countStart) >= TPS_UPDATE_INTERVAL {
		clock.UpdateTps()
	}
}

func (clock *Clock) GetTps() float32 {
	return clock.curTps
	// return 0 // float32(time.Second) / float32(clock.lastTickTime)
}

func (clock *Clock) SetTps(tps int) {
	if tps <= 0 {
		clock.TickTime = 0
	} else {
		clock.TickTime = time.Second / time.Duration(tps)
	}
}

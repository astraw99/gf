// Copyright GoFrame Author(https://goframe.org). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gtimer

import (
	"github.com/gogf/gf/container/gtype"
	"math"
)

// Entry is the timing job.
type Entry struct {
	job       JobFunc      // The job function.
	timer     *Timer       // Belonged timer.
	ticks     int64        // The job runs every ticks.
	times     *gtype.Int   // Limit running times.
	status    *gtype.Int   // Job status.
	singleton *gtype.Bool  // Singleton mode.
	nextTicks *gtype.Int64 // Next run ticks of the job.
}

// JobFunc is the job function.
type JobFunc = func()

// Status returns the status of the job.
func (entry *Entry) Status() int {
	return entry.status.Val()
}

// Run runs the timer job asynchronously.
func (entry *Entry) Run() {
	leftRunningTimes := entry.times.Add(-1)
	if leftRunningTimes < 0 {
		entry.status.Set(StatusClosed)
		return
	}
	// This means it does not limit the running times.
	// I know it's ugly, but it is surely high performance for running times limit.
	if leftRunningTimes < 2000000000 && leftRunningTimes > 1000000000 {
		entry.times.Set(math.MaxInt32)
	}
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if err != panicExit {
					panic(err)
				} else {
					entry.Close()
					return
				}
			}
			if entry.Status() == StatusRunning {
				entry.SetStatus(StatusReady)
			}
		}()
		entry.job()
	}()
}

// doCheckAndRunByTicks checks the if job can run in given timer ticks,
// it runs asynchronously if the given `currentTimerTicks` meets or else
// it increments its ticks and waits for next running check.
func (entry *Entry) doCheckAndRunByTicks(currentTimerTicks int64) {
	// Ticks check.
	if currentTimerTicks < entry.nextTicks.Val() {
		return
	}
	entry.nextTicks.Set(currentTimerTicks + entry.ticks)
	// Perform job checking.
	switch entry.status.Val() {
	case StatusRunning:
		if entry.IsSingleton() {
			return
		}
	case StatusReady:
		if !entry.status.Cas(StatusReady, StatusRunning) {
			return
		}
	case StatusStopped:
		return
	case StatusClosed:
		return
	}
	// Perform job running.
	entry.Run()
}

// SetStatus custom sets the status for the job.
func (entry *Entry) SetStatus(status int) int {
	return entry.status.Set(status)
}

// Start starts the job.
func (entry *Entry) Start() {
	entry.status.Set(StatusReady)
}

// Stop stops the job.
func (entry *Entry) Stop() {
	entry.status.Set(StatusStopped)
}

// Close closes the job, and then it will be removed from the timer.
func (entry *Entry) Close() {
	entry.status.Set(StatusClosed)
}

// Reset reset the job, which resets its ticks for next running.
func (entry *Entry) Reset() {
	entry.nextTicks.Set(entry.timer.ticks.Val() + entry.ticks)
}

// IsSingleton checks and returns whether the job in singleton mode.
func (entry *Entry) IsSingleton() bool {
	return entry.singleton.Val()
}

// SetSingleton sets the job singleton mode.
func (entry *Entry) SetSingleton(enabled bool) {
	entry.singleton.Set(enabled)
}

// Job returns the job function of this job.
func (entry *Entry) Job() JobFunc {
	return entry.job
}

// SetTimes sets the limit running times for the job.
func (entry *Entry) SetTimes(times int) {
	entry.times.Set(times)
}

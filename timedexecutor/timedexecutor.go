package timedexecutor

import (
	"sync"
	"time"

	"github.com/loveandpeople-DAG/goHive/timedqueue"
)

// region TimedExecutor ////////////////////////////////////////////////////////////////////////////////////////////////

// TimedExecutor defines a scheduler that executes tasks in the background at a given time. It does not spawn any
// additional goroutines for each task and executes the tasks sequentially (in each worker).
type TimedExecutor struct {
	workerCount int
	queue       *timedqueue.TimedQueue
	shutdownWG  sync.WaitGroup
}

// New is the constructor for a TimedExecutor that creates a scheduler with a given number of workers that execute the
// scheduled tasks in parallel (whenever they become due).
func New(workerCount int) (timedExecutor *TimedExecutor) {
	timedExecutor = &TimedExecutor{
		workerCount: workerCount,
		queue:       timedqueue.New(),
	}

	timedExecutor.startBackgroundWorkers()

	return
}

// ExecuteAfter executes the given function after the given delay.
func (t *TimedExecutor) ExecuteAfter(f func(), delay time.Duration) *ScheduledTask {
	return t.ExecuteAt(f, time.Now().Add(delay))
}

// ExecuteAt executes the given function at the given time.
func (t *TimedExecutor) ExecuteAt(f func(), time time.Time) *ScheduledTask {
	return t.queue.Add(f, time)
}

// Size returns the amount of jobs that are currently scheduled for execution.
func (t *TimedExecutor) Size() int {
	return t.queue.Size()
}

// WorkerCount returns the amount of background workers that this executor uses.
func (t *TimedExecutor) WorkerCount() int {
	return t.workerCount
}

// Shutdown shuts down the TimedExecutor and waits until the executor has shutdown gracefully.
func (t *TimedExecutor) Shutdown(optionalShutdownFlags ...timedqueue.ShutdownFlag) {
	shutdownFlags := timedqueue.PanicOnModificationsAfterShutdown
	for _, optionalShutdownFlag := range optionalShutdownFlags {
		shutdownFlags |= optionalShutdownFlag
	}

	t.queue.Shutdown(shutdownFlags)

	t.shutdownWG.Wait()
}

// startBackgroundWorkers is an internal utility function that spawns the background workers that execute the queued tasks.
func (t *TimedExecutor) startBackgroundWorkers() {
	for i := 0; i < t.workerCount; i++ {
		t.shutdownWG.Add(1)
		go func() {
			for currentEntry := t.queue.Poll(true); currentEntry != nil; currentEntry = t.queue.Poll(true) {
				currentEntry.(func())()
			}

			t.shutdownWG.Done()
		}()
	}
}

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ScheduledTask ////////////////////////////////////////////////////////////////////////////////////////////////////

// ScheduledTask is
type ScheduledTask = timedqueue.QueueElement

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

// region ShutdownFlag /////////////////////////////////////////////////////////////////////////////////////////////////

// ShutdownFlag defines the type of the optional shutdown flags.
type ShutdownFlag = timedqueue.ShutdownFlag

const (
	// CancelPendingTasks defines a shutdown flag that causes all pending tasks to be canceled.
	CancelPendingTasks = timedqueue.CancelPendingElements

	// IgnorePendingTimeouts defines a shutdown flag, that makes the queue ignore the timeouts of the remaining queued
	// elements. Consecutive calls to Poll will immediately return these elements.
	IgnorePendingTimeouts = timedqueue.IgnorePendingTimeouts
)

// endregion ///////////////////////////////////////////////////////////////////////////////////////////////////////////

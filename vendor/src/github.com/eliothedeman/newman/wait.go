package newman

import "time"

// Waiter provies a method for backing off reading/writing
type Waiter interface {
	Wait()
	Reset()
}

// Backoff is a waiter that sleeps when Wait() is called
type Backoff struct {
	duration time.Duration
}

// Wait sleeps when called, and then doubles the amount of time it will sleep on the next call
func (b *Backoff) Wait() {

	// don't allow sleeping of 0 time. This will not end well
	if b.duration == 0 {
		b.Reset()
	}

	time.Sleep(b.duration)

	// double the sleeping duration for the next call
	b.duration = b.duration * 2
}

// Reset sets the sleep duration back to it's starting value
func (b *Backoff) Reset() {
	b.duration = time.Microsecond
}

// NoopWaiter does nothing when called
type NoopWaiter struct {
}

func (n *NoopWaiter) Wait() {

}

func (n *NoopWaiter) Reset() {

}

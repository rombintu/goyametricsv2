// Package patterns provides various design patterns and concurrency utilities.
package patterns

// Semaphore is a struct that implements a semaphore using a buffered channel.
// It allows controlling the number of concurrent operations by limiting the number of
// goroutines that can access a resource at the same time.
type Semaphore struct {
	semaCh chan struct{} // Buffered channel used to implement the semaphore
}

// NewSemaphore creates a new Semaphore with the specified maximum number of concurrent requests.
//
// Parameters:
// - maxReq: The maximum number of concurrent requests allowed by the semaphore.
//
// Returns:
// - A pointer to the newly created Semaphore.
func NewSemaphore(maxReq int64) *Semaphore {
	return &Semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

// Acquire acquires a semaphore token, blocking if the maximum number of concurrent requests is reached.
// This method should be called before starting a resource-intensive operation.
func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

// Release releases a semaphore token, allowing another goroutine to acquire the token.
// This method should be called after completing a resource-intensive operation.
func (s *Semaphore) Release() {
	<-s.semaCh
}

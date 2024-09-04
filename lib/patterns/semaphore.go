package patterns

type Semaphore struct {
	semaCh chan struct{}
}

func NewSemaphore(maxReq int64) *Semaphore {
	return &Semaphore{
		semaCh: make(chan struct{}, maxReq),
	}
}

func (s *Semaphore) Acquire() {
	s.semaCh <- struct{}{}
}

func (s *Semaphore) Release() {
	<-s.semaCh
}

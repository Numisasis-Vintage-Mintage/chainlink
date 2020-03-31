package util

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// TickerService is an interface that provides
// basic ticking functionality to a service
type TickerService interface {
	Start()
	Stop()
	Tick()
}

// Ticker is the implementation of the TickerService
// interface, managing the lifecycle of a ticking service
type Ticker struct {
	Ticker *time.Ticker
	Name   string
	Impl   TickerService
	Done   chan struct{}
	Exited chan struct{}
}

// Start will start the ticker service
func (t *Ticker) Start() {
	go t.tick()
}

// Stop will stop the ticket service, waiting
// for any any tick to first complete
func (t *Ticker) Stop() {
	close(t.Done)
	t.WithService().Infof("Waiting for service to exit")
	<-t.Exited
	t.WithService().Infof("Safely exited")
}

// WithService will return a logrus log entry with the service name
// specified as a field
func (t *Ticker) WithService() *log.Entry {
	return log.WithField("service", t.Name)
}

func (t *Ticker) tick() {
	t.WithService().Infof("Running initial service tick")
	t.Impl.Tick()
	t.WithService().Infof("Initial tick complete")

	defer close(t.Exited)
	for {
		select {
		case <-t.Done:
			return
		case <-t.Ticker.C:
			t.WithService().Debug("Service tick")
			t.Impl.Tick()
			t.WithService().Debug("Tick complete")
		}
	}
}

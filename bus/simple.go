package bus

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/InnovaCo/broforce/logger"
)

func init() {
	registry([]string{".*"}, adapter(&simpleAdapter{}))
}

func SafeHandler(h Handler, sp SafeParams) Handler {
	return func(e Event) error {
		defer timeTrack(time.Now(), runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name())
		for {
			if err := h(e); err != nil {
				logger.Log.Error(err)
				if sp.Retry <= 0 {
					return err
				} else {
					logger.Log.Debug("Retry")
					sp.Retry--
					time.Sleep(sp.Delay)
				}
			} else {
				break
			}
		}
		return nil
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	logger.Log.Debugf("func: %s, work time: %s", name, elapsed)
}

type simpleAdapter struct {
	subs map[string][]Handler
	lock sync.Mutex
}

func (p *simpleAdapter) Run() error {
	p.lock = sync.Mutex{}
	p.subs = make(map[string][]Handler)
	return nil
}

func (p *simpleAdapter) Publish(e Event) error {
	if _, ok := p.subs[e.Subject]; !ok {
		return fmt.Errorf("subs for %s empty", e.Subject)
	}
	logger.Log.Debug("<-->", e.Subject)
	for _, h := range p.subs[e.Subject] {
		go h(e)
	}
	return nil
}

func (p *simpleAdapter) Subscribe(subject string, h Handler) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.subs[subject]; !ok {
		p.subs[subject] = make([]Handler, 0)
	}
	p.subs[subject] = append(p.subs[subject], SafeHandler(h, SafeParams{Retry: 1, Delay: time.Duration(1)}))
}

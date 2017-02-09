package bus

import (
	"fmt"
	"sync"
	"time"

	"github.com/InnovaCo/broforce/logger"
)

func init() {
	registry([]string{".*"}, adapter(&simpleAdapter{}))
}

func SafeHandler(h Handler, sp SafeParams) Handler {
	return func(e Event, ctx Context) error {
		updateContext(e, &ctx)
		defer timeTrack(time.Now(), ctx)
		for {
			if err := h(e, ctx); err != nil {
				ctx.Log.Error(err)
				if sp.Retry <= 0 {
					return err
				} else {
					ctx.Log.Debug("Retry")
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

func updateContext(e Event, ctx *Context) {
	ctx.Log = logger.Logger4Handler(ctx.Name, e.Trace)
}

func timeTrack(start time.Time, ctx Context) {
	elapsed := time.Since(start)
	ctx.Log.Debugf("func: %s, work time: %s", ctx.Name, elapsed)
}

type simpleAdapter struct {
	subs map[string][]Context
	lock sync.Mutex
}

func (p *simpleAdapter) Run() error {
	p.lock = sync.Mutex{}
	p.subs = make(map[string][]Context)
	return nil
}

func (p *simpleAdapter) Publish(e Event) error {
	if _, ok := p.subs[e.Subject]; !ok {
		return fmt.Errorf("subs for %s empty", e.Subject)
	}
	for _, ctx := range p.subs[e.Subject] {
		go ctx.Func(e, ctx)
	}
	return nil
}

func (p *simpleAdapter) Subscribe(subject string, ctx Context) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.subs[subject]; !ok {
		p.subs[subject] = make([]Context, 0)
	}
	ctx.Func = SafeHandler(ctx.Func, SafeParams{Retry: 0, Delay: time.Duration(1)})
	p.subs[subject] = append(p.subs[subject], ctx)
}

package bus

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"

	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

var (
	once        sync.Once
	instance    *EventsBus
	busAdapters = make([]*adapterConfig, 0)
)

type Context struct {
	Func   Handler
	Name   string
	Log    *log.Entry
	Config config.ConfigData
	Bus    *EventsBus
}

type SafeParams struct {
	Retry int
	Delay time.Duration
}

type Task interface {
	Run(ctx Context) error
}

type Handler func(e Event, ctx Context) error

type EventsBus struct {
}

type adapter interface {
	Run(cfg config.ConfigData) error
	Publish(e Event) error
	Subscribe(subject string, ctx Context)
}

type adapterConfig struct {
	Name       string
	EventTypes []*regexp.Regexp
	Adapter    adapter
}

func registry(name string, adap adapter) {
	for _, acfg := range busAdapters {
		if strings.Compare(acfg.Name, name) == 0 {
			logger.Log.Errorf("Adapter %s rewrite", name)
			acfg.Adapter = adap
			return
		}
	}
	busAdapters = append(busAdapters, &adapterConfig{
		Name:       name,
		EventTypes: make([]*regexp.Regexp, 0),
		Adapter:    adap})
	return
}

func GetNameAdapters() []string {
	result := make([]string, 0)
	for _, acfg := range busAdapters {
		result = append(result, acfg.Name)
	}
	return result
}

func NewUUID() string {
	return uuid.NewV4().String()
}

func SafeRun(r func(ctx Context) error, sp SafeParams) func(ctx Context) error {
	return func(ctx Context) error {
		for {
			if err := r(ctx); err != nil {
				logger.Log.Error(err)
				if sp.Retry <= 0 {
					return err
				} else {
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

func withContextLogger(h Handler) Handler {
	return func(e Event, ctx Context) error {
		ctx.Log = logger.Logger4Handler(ctx.Name, e.Trace)
		defer timeTrack(time.Now(), ctx)
		return h(e, ctx)
	}
}

func timeTrack(start time.Time, ctx Context) {
	elapsed := time.Since(start)
	ctx.Log.Debugf("func: %s, work time: %s", ctx.Name, elapsed)
}

func New(cfg config.ConfigData) *EventsBus {
	once.Do(func() {
		for _, acfg := range busAdapters {
			for _, et := range cfg.GetArrayString(fmt.Sprintf("%s.event-types", acfg.Name)) {
				if r, err := regexp.Compile(et); err == nil {
					acfg.EventTypes = append(acfg.EventTypes, r)
				} else {
					logger.Log.Errorf("Error: event type %s for adapter % not compile", et, acfg.Name)
				}
			}
			if err := acfg.Adapter.Run(cfg.Get(acfg.Name)); err != nil {
				logger.Log.Errorf("Error: %v", err)
			}
		}
		instance = &EventsBus{}
	})
	return instance
}

func (p *EventsBus) Publish(e Event) error {
	for _, acfg := range busAdapters {
		for _, et := range acfg.EventTypes {
			if len(et.FindAllString(e.Subject, -1)) == 1 {
				return acfg.Adapter.Publish(e)
			}
		}
	}
	return nil
}

func (p *EventsBus) Subscribe(subject string, ctx Context) {
	for _, acfg := range busAdapters {
		for _, et := range acfg.EventTypes {
			if len(et.FindAllString(subject, -1)) == 1 {
				ctx.Func = withContextLogger(ctx.Func)
				if ctx.Log == nil {
					ctx.Log = logger.Logger4Handler(ctx.Name, "")
				}
				acfg.Adapter.Subscribe(subject, ctx)
			}
		}
	}
}

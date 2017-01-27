package tasks

import (
	"time"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

func init() {
	//registry("timer", bus.Task(&timer{}))
}

type Tact struct {
	Number int64 `json:"number"`
}

type timer struct {
	interval time.Duration
}

func (p *timer) handler(e bus.Event) error {
	tact := Tact{}
	if err := bus.Encoder(e.Data, &tact, e.Coding); err != nil {
		return err
	}

	logger.Log.Debugf("Tact: %d", tact.Number)
	return nil
}

func (p *timer) Run(eventBus *bus.EventsBus, cfg config.ConfigData) error {
	p.interval = time.Duration(cfg.GetIntOr("interval", 1)) * time.Second
	eventBus.Subscribe(bus.TimerEvent, p.handler)

	i := int64(0)
	e := bus.Event{
		Trace:   bus.NewUUID(),
		Subject: bus.TimerEvent,
		Coding:  bus.JsonCoding}

	tact := Tact{}
	for {
		tact.Number, i = i, i+1

		if err := bus.Coder(&e, tact); err != nil {
			logger.Log.Error(err)
		}
		if err := eventBus.Publish(e); err != nil {
			logger.Log.Error(err)
		}
		time.Sleep(p.interval)
	}
	logger.Log.Debug("timeSensor Complete")
	return nil
}

package tasks

import (
	"time"

	"github.com/InnovaCo/broforce/bus"
)

func init() {
	//registry("timer", bus.Task(&timer{}))
}

//config section
//
//timer:
//  interval: 10
//

type Tact struct {
	Number int64 `json:"number"`
}

type timer struct {
	interval time.Duration
}

func (p *timer) handler(e bus.Event, ctx bus.Context) error {
	tact := Tact{}
	if err := e.Unmarshal(&tact); err != nil {
		return err
	}

	ctx.Log.Debugf("Tact: %d", tact.Number)
	return nil
}

func (p *timer) Run(ctx bus.Context) error {
	p.interval = time.Duration(ctx.Config.GetIntOr("interval", 1)) * time.Second
	ctx.Bus.Subscribe(bus.TimerEvent, bus.Context{
		Func:   p.handler,
		Name:   "TimerHandler",
		Bus:    ctx.Bus,
		Config: ctx.Config})
	i := int64(0)
	tact := Tact{}
	for {
		tact.Number, i = i, i+1
		uuid := bus.NewUUID()
		if event, err := bus.NewEventWithData(uuid, bus.TimerEvent, bus.JsonCoding, tact); err != nil {
			ctx.Log.Error(err)
		} else {
			ctx.Log.Debugf("Push: %s", uuid)

			if err := ctx.Bus.Publish(*event); err != nil {
				ctx.Log.Error(err)
			}
		}
		time.Sleep(p.interval)
	}

	ctx.Log.Debug("timer Complete")

	return nil
}

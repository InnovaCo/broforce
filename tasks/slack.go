package tasks

import (
	"fmt"
	"strings"

	"github.com/nlopes/slack"

	"github.com/InnovaCo/broforce/bus"
)

func init() {
	registry("slackSensor", bus.Task(&sensorSlack{}))
}

type slackMessage slack.Msg

type sensorSlack struct {
	bus    *bus.EventsBus
	client *slack.Client
	user   *slack.User
}

func (p *sensorSlack) messageEvent(msg *slack.MessageEvent, ctx *bus.Context) error {
	if strings.Compare(msg.User, p.user.ID) == 0 {
		ctx.Log.Debugf("Ignore message: '%s'", msg.Text)
		return nil
	}

	ctx.Log.Debugf("User: %s, channel: %s, message: '%s'", msg.User, msg.Channel, msg.Text)

	event := bus.Event{
		Trace:   bus.NewUUID(),
		Subject: bus.SlackMsgEvent,
		Coding:  bus.JsonCoding}

	if err := bus.Coder(&event, msg.Msg); err == nil {
		if err := p.bus.Publish(event); err != nil {
			ctx.Log.Error(err)
		}
	} else {
		return err
	}
	return nil
}

func (p *sensorSlack) postMessage(e bus.Event, ctx bus.Context) error {
	msg := slackMessage{}
	if err := bus.Encoder(e.Data, &msg, e.Coding); err != nil {
		return err
	}

	ctx.Log.Debugf("%s: '%s'", msg.Channel, msg.Text)

	params := slack.PostMessageParameters{
		AsUser:   true,
		Username: p.user.ID}

	if len(msg.Attachments) != 0 {
		params.Attachments = msg.Attachments
	}

	_, _, err := p.client.PostMessage(msg.Channel, msg.Text, params)
	return err
}

func (p *sensorSlack) Run(eventBus *bus.EventsBus, ctx bus.Context) error {
	p.bus = eventBus
	p.client = slack.New(ctx.Config.GetStringOr("token", ""))

	if user, err := p.client.GetUserInfo(ctx.Config.GetString("username")); err != nil {
		return err
	} else {
		p.user = user
	}

	ctx.Log.Infof("Ignore user: %s", p.user.ID)

	rtm := p.client.NewRTM()
	go rtm.ManageConnection()

	p.bus.Subscribe(bus.SlackPostEvent, bus.Context{Func: p.postMessage, Name: "SlackHandler"})

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if err := p.messageEvent(ev, &ctx); err != nil {
					ctx.Log.Error(err)
				}

			case *slack.RTMError:
				ctx.Log.Error(ev.Error())

			case *slack.InvalidAuthEvent:
				return fmt.Errorf("Invalid credentials")

			}
		}
	}
	return nil
}

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

//config section
//
//slackSensor:
//  username: "UUID user"
//  token: TOKEN
//

type slackMessage slack.Msg

type sensorSlack struct {
	client *slack.Client
	user   *slack.User
}

func (p *sensorSlack) messageEvent(msg *slack.MessageEvent, ctx *bus.Context) error {
	if strings.Compare(msg.User, p.user.ID) == 0 {
		//ctx.Log.Debugf("Ignore message: '%s'", msg.Text)
		return nil
	}

	//ctx.Log.Debugf("User: %s, channel: %s, message: '%s'", msg.User, msg.Channel, msg.Text)

	uuid := bus.NewUUID()
	if event, err := bus.NewEventWithData(uuid, bus.SlackMsgEvent, bus.JsonCoding, msg.Msg); err != nil {
		return err
	} else {
		ctx.Log.Debugf("Push: %s", uuid)
		return ctx.Bus.Publish(*event)
	}
}

func (p *sensorSlack) postMessage(e bus.Event, ctx bus.Context) error {
	msg := slackMessage{}
	if err := e.Unmarshal(&msg); err != nil {
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

func (p *sensorSlack) Run(ctx bus.Context) error {
	p.client = slack.New(ctx.Config.GetStringOr("token", ""))

	if user, err := p.client.GetUserInfo(ctx.Config.GetString("username")); err != nil {
		return err
	} else {
		p.user = user
	}

	ctx.Log.Infof("Ignore user: %s", p.user.ID)

	rtm := p.client.NewRTM()
	go rtm.ManageConnection()

	ctx.Bus.Subscribe(bus.SlackPostEvent, bus.Context{
		Func:   p.postMessage,
		Name:   "SlackHandler",
		Bus:    ctx.Bus,
		Config: ctx.Config})

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

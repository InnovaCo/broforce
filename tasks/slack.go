package tasks

import (
	"fmt"
	"strings"

	"github.com/nlopes/slack"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
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

func (p *sensorSlack) messageEvent(msg *slack.MessageEvent) error {
	logger.Log.Debugf("--> %v %v %v", msg.User, msg.Channel, msg.Text)

	if strings.Compare(msg.User, p.user.ID) == 0 {
		logger.Log.Debug("Ignore message")
		return nil
	}

	event := bus.Event{Subject: bus.SlackMsgEvent, Coding: bus.JsonCoding}

	if err := bus.Coder(&event, msg.Msg); err == nil {
		if err := p.bus.Publish(event); err != nil {
			logger.Log.Error(err)
		}
	} else {
		return err
	}
	return nil
}

func (p *sensorSlack) postMessage(e bus.Event) error {
	msg := slackMessage{}
	if err := bus.Encoder(e.Data, &msg, e.Coding); err != nil {
		return err
	}

	logger.Log.Debug("<--", msg.Channel, msg.Text)

	params := slack.PostMessageParameters{
		AsUser:   true,
		Username: p.user.ID}

	if len(msg.Attachments) != 0 {
		params.Attachments = msg.Attachments
	}

	_, _, err := p.client.PostMessage(msg.Channel, msg.Text, params)
	return err
}

func (p *sensorSlack) Run(eventBus *bus.EventsBus, cfg config.ConfigData) error {
	logger.Log.Debug(cfg.String())
	p.bus = eventBus
	p.client = slack.New(cfg.GetStringOr("token", ""))

	if user, err := p.client.GetUserInfo(cfg.GetString("username")); err != nil {
		return err
	} else {
		p.user = user
	}

	logger.Log.Infof("Ignore user: %s", p.user.ID)

	rtm := p.client.NewRTM()
	go rtm.ManageConnection()

	p.bus.Subscribe(bus.SlackPostEvent, p.postMessage)

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.MessageEvent:
				if err := p.messageEvent(ev); err != nil {
					logger.Log.Error(err)
				}

			case *slack.RTMError:
				logger.Log.Error(ev.Error())

			case *slack.InvalidAuthEvent:
				return fmt.Errorf("Invalid credentials")

			}
		}
	}
	return nil
}

package tasks

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/hashicorp/consul/api"

	"github.com/InnovaCo/broforce/bus"
	"github.com/InnovaCo/broforce/config"
	"github.com/InnovaCo/broforce/logger"
)

const (
	dataPrefix     = "services/data"
	outdatedPrefix = "services/outdated"
	loopInterval   = 10
)

type outdatedEvent struct {
	EndOfLife int64  `json:"endOfLife"`
	Key       string `json:"key"`
	Address   string `json:"address"`
}

func init() {
	registry("consulSensor", bus.Task(&consulSensor{}))
	registry("outdated", bus.Task(&outdatedConsul{}))
}

type consulSensor struct {
	clientsPool map[string]*api.Client
}

func (p *consulSensor) prepareConfig(cfg config.ConfigData) []*api.Config {
	dc := make([]*api.Config, 0)
	for _, address := range cfg.GetArray("consul") {
		c := api.DefaultConfig()
		logger.Log.Debug(address.String())
		c.Address = address.String()
		dc = append(dc, c)
	}
	return dc
}

func (p *consulSensor) Run(eventBus *bus.EventsBus, cfg config.ConfigData) error {
	logger.Log.Debug(cfg.String())

	p.clientsPool = make(map[string]*api.Client)

	for _, c := range p.prepareConfig(cfg) {
		client, err := api.NewClient(c)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		p.clientsPool[c.Address] = client
	}
	for {
		for address, client := range p.clientsPool {
			kv := client.KV()
			pairs, _, err := kv.List(outdatedPrefix, nil)
			if err != nil {
				logger.Log.Error(err)
				continue
			}
			for _, key := range pairs {
				logger.Log.Debugf("KV: %v=%v", string(key.Key), string(key.Value))
				outdated := outdatedEvent{EndOfLife: -1}
				if err := json.Unmarshal(key.Value, &outdated); err != nil {
					logger.Log.Error(err)
				}
				if outdated.EndOfLife == -1 {
					continue
				}
				if outdated.EndOfLife < time.Now().UnixNano()/int64(time.Millisecond) {
					outdated.Key = strings.Replace(key.Key, outdatedPrefix+"/", "", 1)
					outdated.Address = address
					e := bus.Event{Subject: bus.OutdatedEvent, Coding: bus.JsonCoding}
					if err := bus.Coder(&e, outdated); err == nil {
						if err := eventBus.Publish(e); err != nil {
							logger.Log.Error(err)
						}
					} else {
						logger.Log.Error(err)
					}
				} else {
					logger.Log.Debugf("outdated delta: %v", outdated.EndOfLife-time.Now().UnixNano()/int64(time.Millisecond))
				}
			}
		}
		time.Sleep(loopInterval * time.Second)
	}
	logger.Log.Debug("consulSensor Complete")
	return nil
}

type outdatedConsul struct {
	bus *bus.EventsBus
}

func (p *outdatedConsul) handler(e bus.Event) error {
	event := outdatedEvent{}
	if err := bus.Encoder(e.Data, &event, e.Coding); err != nil {
		return err
	}

	logger.Log.Debug(event)

	conf := api.DefaultConfig()
	conf.Address = event.Address
	client, err := api.NewClient(conf)
	if err != nil {
		return err
	}
	kv := client.KV()
	pairs, _, err := kv.List(dataPrefix+"/"+event.Key+"/", nil)
	if err != nil {
		return err
	}

	if len(pairs) == 0 {
		logger.Log.Infof("key %s empty, delete key: %s", dataPrefix+"/"+event.Key+"/", outdatedPrefix+"/"+event.Key)

		if _, err := kv.Delete(outdatedPrefix+"/"+event.Key, nil); err != nil {
			return err
		}
		return nil
	}

	logger.Log.Debug(pairs)

	serveEvent := bus.Event{Subject: bus.ServeCmdWithDataEvent, Coding: bus.JsonCoding}

	for _, key := range pairs {
		logger.Log.Debugf("KV: %v=%v", string(key.Key), string(key.Value))
		g, err := gabs.ParseJSON(key.Value)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		g.Set("true", "purge")
		plugin := strings.Split(key.Key, "/")
		params := serveParams{
			Vars:     map[string]string{"purge": "true"},
			Plugin:   plugin[len(plugin)-1],
			Manifest: g.Bytes()}
		if err := bus.Coder(&serveEvent, params); err != nil {
			logger.Log.Error(err)
			continue
		}
		if err := p.bus.Publish(serveEvent); err != nil {
			logger.Log.Error(err)
		}
	}
	return nil
}

func (p *outdatedConsul) Run(eventBus *bus.EventsBus, cfg config.ConfigData) error {
	logger.Log.Debug(cfg.String())

	p.bus = eventBus
	p.bus.Subscribe(bus.OutdatedEvent, p.handler)
	return nil
}

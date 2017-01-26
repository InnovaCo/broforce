package bus

import (
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/nats-io/go-nats"

	"github.com/InnovaCo/broforce/logger"
)

func init() {
	//registry([]string{}, adapter(&natsAdapter{}))
}

type natsAdapter struct {
	url  string
	conn *nats.Conn
	subs map[string][]Handler
	lock sync.Mutex
}

func (p *natsAdapter) Run() error {
	p.lock = sync.Mutex{}
	p.subs = make(map[string][]Handler)
	var err error
	p.runServer()

	for i := 0; i < 5; i = i + 1 {
		p.conn, err = nats.Connect("nats://localhost:4222")
		if err != nil {
			logger.Log.Errorf("Can't connect: %v\n", err)
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	return p.conn.LastError()
}

func (p *natsAdapter) runServer() error {
	cmd := exec.Command("gnatsd", "-a", "localhost")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		logger.Log.Errorf("Run server error: %v\n", err)
		return err
	}
	return nil
}

func (p *natsAdapter) handler(msg *nats.Msg) {
	logger.Log.Debug("--> ", msg.Subject)
	if _, ok := p.subs[msg.Subject]; !ok {
		return
	}

	for _, t := range p.subs[msg.Subject] {
		go t(Event{Subject: msg.Subject, Data: msg.Data})
	}
}

func (p *natsAdapter) Publish(e Event) error {
	logger.Log.Debug("<-- ", e.Subject)
	return p.conn.Publish(e.Subject, e.Data)
}

func (p *natsAdapter) Subscribe(subject string, h Handler) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if _, ok := p.subs[subject]; !ok {
		p.subs[subject] = make([]Handler, 0)
		logger.Log.Debugf("subs %s", subject)
		if s, err := p.conn.Subscribe(subject, p.handler); err != nil {
			logger.Log.Error(err)
		} else {
			logger.Log.Debug(s.Subject)
		}
		p.conn.Flush()
	}
	p.subs[subject] = append(p.subs[subject], h)

	logger.Log.Debug(p.subs)
}

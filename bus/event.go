package bus

import (
	"encoding/json"
)

type Event struct {
	Trace   string
	Subject string
	Coding  string
	Data    []byte
}

func NewEvent(trace string, subject string, coding string) *Event {
	return &Event{Trace: trace,
		Subject: subject,
		Coding:  coding,
		Data:    make([]byte, 0)}
}

func NewEventWithData(trace string, subject string, coding string, data interface{}) (*Event, error) {
	event := Event{Trace: trace,
		Subject: subject,
		Coding:  coding,
		Data:    make([]byte, 0)}
	return &event, event.Marshal(data)
}

func (p *Event) Marshal(d interface{}) error {
	var err error
	switch p.Coding {
	case JsonCoding:
		if p.Data, err = json.Marshal(d); err != nil {
			return err
		}
	}
	return nil
}

func (p *Event) Unmarshal(v interface{}) error {
	switch p.Coding {
	case JsonCoding:
		if err := json.Unmarshal(p.Data, v); err != nil {
			return err
		}
	}
	return nil
}

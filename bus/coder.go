package bus

import (
	"encoding/json"
)

func Encoder(d []byte, v interface{}, conding string) error {
	switch conding {
	case JsonCoding:
		if err := json.Unmarshal(d, v); err != nil {
			return err
		}
	}
	return nil
}

func Coder(e *Event, d interface{}) error {
	var err error
	switch e.Coding {
	case JsonCoding:
		if e.Data, err = json.Marshal(d); err != nil {
			return err
		}
	}
	return nil
}

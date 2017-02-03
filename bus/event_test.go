package bus

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvent(t *testing.T) {
	type data struct {
		Param1 int               `json:"param1"`
		Param2 []string          `json:"param2"`
		Param3 map[string]string `json:"param3"`
	}

	t.Run("Marshal", func(t *testing.T) {
		e := Event{Subject: UnknownEvent, Coding: JsonCoding}
		if err := e.Marshal(data{
			Param1: 1,
			Param2: []string{"val1", "val2"},
			Param3: map[string]string{"pp3": "val1"}}); err != nil {
			t.Error(err)
			t.Fail()
		}

		result := `{"param1":1,"param2":["val1","val2"],"param3":{"pp3":"val1"}}`

		if strings.Compare(string(e.Data), result) != 0 {
			t.Errorf("%s != %s", string(e.Data), result)
			t.Fail()
		}
	})

	t.Run("Unmarshal", func(t *testing.T) {
		e := Event{
			Subject: UnknownEvent,
			Coding:  JsonCoding,
			Data:    []byte(`{"param1":1,"param2":["val1","val2"],"param3":{"pp3":"val1"}}`)}

		d := data{}

		if err := e.Unmarshal(&d); err != nil {
			t.Error(err)
			t.Fail()
		}

		if !reflect.DeepEqual(d.Param1, 1) {
			t.Errorf("%v != %v", d.Param1, 1)
			t.Fail()
		}
		if !reflect.DeepEqual(d.Param2, []string{"val1", "val2"}) {
			t.Errorf("%v != %v", d.Param2, []string{"val1", "val2"})
			t.Fail()
		}
		if !reflect.DeepEqual(d.Param3, map[string]string{"pp3": "val1"}) {
			t.Errorf("%v != %v", d.Param3, map[string]string{"pp3": "val1"})
			t.Fail()
		}
	})
}

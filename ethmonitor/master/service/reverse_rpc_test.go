package service

import (
	"encoding/json"
	"reflect"
	"testing"
)

type test_msg struct {
	Msg string
}

func TestInterfaceReceiver(t *testing.T) {
	o := test_msg{
		Msg: "hello",
	}
	b, _ := json.Marshal(&o)

	var r interface{}
	r = reflect.New(reflect.TypeOf(o)).Interface()
	err := json.Unmarshal(b, r)
	if err != nil {
		t.Fatal("interface receiver failed")
	}
}

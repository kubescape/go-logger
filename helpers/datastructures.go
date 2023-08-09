package helpers

import (
	"strings"
	"time"
)

const InvalidUtf8ReplacementString = "\uFFFD"

var _ IDetails = (*StringObj)(nil)

type StringObj struct {
	key   string
	value string
}

var _ IDetails = (*ErrorObj)(nil)

type ErrorObj struct {
	key   string
	value error
}

var _ IDetails = (*IntObj)(nil)

type IntObj struct {
	key   string
	value int
}

var _ IDetails = (*InterfaceObj)(nil)

type InterfaceObj struct {
	key   string
	value interface{}
}

func Error(e error) *ErrorObj     { return &ErrorObj{key: "error", value: e} }
func Int(k string, v int) *IntObj { return &IntObj{key: k, value: v} }
func String(k, v string) *StringObj {
	return &StringObj{key: k, value: strings.ToValidUTF8(v, InvalidUtf8ReplacementString)}
}
func Interface(k string, v interface{}) *InterfaceObj { return &InterfaceObj{key: k, value: v} }
func Time() *StringObj {
	return &StringObj{key: "time", value: time.Now().Format("2006-01-02 15:04:05")}
}

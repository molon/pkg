package putil

import (
	jsoniter "github.com/json-iterator/go"
)

func MustMarshalToString(v interface{}) string {
	jsn, _ := jsoniter.MarshalToString(v)
	return jsn
}

type JsonMap map[string]interface{}

func (m JsonMap) String() string {
	jsn, _ := jsoniter.MarshalToString(m)
	return jsn
}

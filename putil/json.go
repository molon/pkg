package putil

import (
	jsoniter "github.com/json-iterator/go"
)

func MustMarshalToJson(v interface{}) string {
	jsn, _ := jsoniter.MarshalToString(v)
	return jsn
}

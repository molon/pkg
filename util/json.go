package util

import (
	jsoniter "github.com/json-iterator/go"
)

type JsonMap map[string]interface{}

func (m JsonMap) String() string {
	jsn, _ := jsoniter.MarshalToString(m)
	return jsn
}

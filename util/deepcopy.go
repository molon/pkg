package util

import (
	"bytes"
	"encoding/gob"
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
)

func GobDeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func JsoniterDeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := jsoniter.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return jsoniter.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

func JsonDeepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

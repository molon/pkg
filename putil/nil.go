package putil

import "reflect"

func IsNil(pb interface{}) bool {
	v := reflect.ValueOf(pb)
	return pb == nil || (v.Kind() == reflect.Ptr && v.IsNil())
}

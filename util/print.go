package util

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
)

func PrintJson(as ...interface{}) {
	for idx := 0; idx < len(as); idx++ {
		a := as[idx]
		jsn, _ := jsoniter.MarshalIndent(a, "", "    ")
		fmt.Println(string(jsn))
		if idx != len(as)-1 {
			fmt.Println("---------")
		}
	}
}

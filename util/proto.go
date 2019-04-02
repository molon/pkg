package util

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func ProtoToJSONString(pb proto.Message) string {
	marshaler := &jsonpb.Marshaler{
		OrigName: true,
		Indent:   "    ",
	}
	str, _ := marshaler.MarshalToString(pb)
	return str
}

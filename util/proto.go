package util

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

func ProtoToJSONStringForPrint(pb proto.Message) string {
	marshaler := &jsonpb.Marshaler{
		OrigName: true,
		Indent:   "    ",
	}
	str, _ := marshaler.MarshalToString(pb)
	return str
}

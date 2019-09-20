package putil

import (
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
)

func ProtoToJSONStringForPrint(pb proto.Message) string {
	marshaler := &jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
		Indent:       "    ",
	}
	str, _ := marshaler.MarshalToString(pb)
	return str
}

func ToTimestampProto(t time.Time) (*tspb.Timestamp, error) {
	if t.IsZero() {
		return nil, nil
	}
	return ptypes.TimestampProto(t)
}

func FromTimestampProto(ts *tspb.Timestamp) (time.Time, error) {
	if ts == nil {
		return time.Time{}, nil
	}
	return ptypes.Timestamp(ts)
}

func MarshalAny(pb proto.Message) (*any.Any, error) {
	value, err := proto.Marshal(pb)
	if err != nil {
		return nil, err
	}
	return &any.Any{TypeUrl: "/" + proto.MessageName(pb), Value: value}, nil
}

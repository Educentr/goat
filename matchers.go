package goat

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// ProtoMatcher implements the gomock.Matcher interface
type ProtoMatcher struct {
	Msg proto.Message
}

func (r ProtoMatcher) Matches(msg interface{}) bool {
	m, ok := msg.(proto.Message)
	if !ok {
		return false
	}

	return proto.Equal(m, r.Msg)
}

func (r ProtoMatcher) String() string {
	return fmt.Sprintf("is %s", r.Msg)
}

package api

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// api.Process
func (p *Process) Address() string {
	return fmt.Sprintf("%v:%v", p.Ip, p.Port)
}

func GetRawAddress(address string) (string, string) {
	splitted := strings.Split(address, ":")
	return splitted[0], splitted[1]
}

func IsSameProcess(p1 *Process, p2 *Process) bool {
	if p1 == nil || p2 == nil {
		return false
	}
	return p1.Ip == p2.Ip && p1.Port == p2.Port && p1.JoinTime.AsTime() == p2.JoinTime.AsTime()
}

// api.Sequence
func (s *Sequence) Less(other *Sequence) bool {
	// compare timestamp first
	if s.Time.AsTime().Equal(other.Time.AsTime()) {
		// if the time is the same, compare the process
		return s.Count < other.Count
	}
	return s.Time.AsTime().Before(other.Time.AsTime())
}

func (s *Sequence) Equal(other *Sequence) bool {
	return s.Time.AsTime() == other.Time.AsTime() && s.Count == other.Count
}

// timestamppb.Timestamp
func CurrentTimestamp() *timestamppb.Timestamp {
	return timestamppb.New(time.Now())
}

func TimeSince(startTime time.Time) *timestamppb.Timestamp {
	return &timestamppb.Timestamp{
		Seconds: int64(time.Since(startTime).Seconds()),
		Nanos:   int32(time.Since(startTime).Nanoseconds()),
	}
}

/*
 * Marshal a Metadata struct into a byte array
 *
 * @type T: type of message
 * @param messageType: type of message
 * @param metadata: metadata to marshal
 * @return []byte: marshalled metadata
 * @return error: raise error if marshalling fails
 */
func MarshalMeta(messageType MessageType, message isMetadata_Message) ([]byte, error) {
	meta := &Metadata{
		Type:    messageType,
		Message: message,
	}

	return proto.Marshal(meta)
}

/*
 * Unmarshal a message into a Metadata struct given the byte array
 *
 * @param message: byte array to unmarshal
 * @return *api.Metadata: unmarshalled metadata
 * @return error: raise error if unmarshalling fails
 */
func UnmarshalMeta(message []byte) (*Metadata, error) {
	meta := &Metadata{}
	err := proto.Unmarshal(message, meta)

	return meta, err
}

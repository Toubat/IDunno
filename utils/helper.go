package utils

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"mp4/api"

	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const OK = "OK"
const FAIL = "FAIL"
const EMPTY_STRING = ""

var LAST_SECONDS = math.MaxFloat64
var DEFAULT_TIMESTAMP = timestamppb.Timestamp{
	Seconds: 0,
	Nanos:   0,
}

var source = rand.NewSource(time.Now().UnixNano())
var r = rand.New(source)

/*
 * Drop a message with a given probability
 *
 * @param prob: probability of dropping a message
 * @param callback: callback function to write a message
 * @return int: number of bytes written
 * @return error: raise error if writing fails
 */
func WithDropProb(prob float64, callback func() (int, error)) (int, error) {
	if r.Float64() <= prob {
		return 0, fmt.Errorf("network packet dropped, fail probability: %v", prob)
	}
	return callback()
}

func ConcatFilename(filename string, seq *api.Sequence) string {
	return fmt.Sprintf("[%v][%v][%d]", filename, seq.GetTime().AsTime(), seq.GetCount())
}

func Hash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32())
}

func CreateTempFilename() string {
	return fmt.Sprintf("[%v]", api.CurrentTimestamp())
}

func CreateId(prefix string) string {
	return fmt.Sprintf("%v:%v", prefix, api.CurrentTimestamp().Seconds)
}

func SetLastSeconds(seconds float64) {
	LAST_SECONDS = seconds
}

func ConvertTo64(ar []float32) []float64 {
	newar := make([]float64, len(ar))
	var v float32
	var i int
	for i, v = range ar {
		newar[i] = float64(v)
	}
	return newar
}

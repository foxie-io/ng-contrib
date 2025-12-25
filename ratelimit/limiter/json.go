package limiter

import (
	"encoding/json"
	"fmt"
)

type JSONSerializer interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(data []byte, v interface{}) error
}

// Default JSON serializer uses encoding/json
var DefaultSerializer JSONSerializer = &stdJSON{}

type stdJSON struct{}

func (s *stdJSON) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (s *stdJSON) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// LimiterJSON is a wrapper to marshal/unmarshal any Limiter implementation.
type LimiterJSON struct {
	Limiter Limiter
}

func (w LimiterJSON) MarshalJSON() ([]byte, error) {
	var typeName string
	switch w.Limiter.Algorithm() {
	case AlgorithmTokenBucket:
		typeName = "tokenBucket"
	case AlgorithmFixedWindow:
		typeName = "fixedWindow"
	case AlgorithmSlidingWindow:
		typeName = "slidingWindow"
	default:
		return nil, ErrUnknownLimiter
	}

	return DefaultSerializer.Marshal(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: typeName,
		Data: w.Limiter,
	})
}

func (w *LimiterJSON) UnmarshalJSON(data []byte) error {
	var wrapper struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := DefaultSerializer.Unmarshal(data, &wrapper); err != nil {
		return err
	}

	var l Limiter
	switch wrapper.Type {
	case "tokenBucket":
		var tb TokenBucket
		if err := DefaultSerializer.Unmarshal(wrapper.Data, &tb); err != nil {
			return err
		}
		l = &tb
	case "fixedWindow":
		var fw FixedWindow
		if err := DefaultSerializer.Unmarshal(wrapper.Data, &fw); err != nil {
			return err
		}
		l = &fw
	case "slidingWindow":
		var sw SlidingWindow
		if err := DefaultSerializer.Unmarshal(wrapper.Data, &sw); err != nil {
			return err
		}
		l = &sw
	default:
		return ErrUnknownLimiter
	}

	w.Limiter = l
	return nil
}

// ErrUnknownLimiter is returned if the type is not recognized
var ErrUnknownLimiter = fmt.Errorf("unknown limiter type")

package msg

import (
	"bytes"
	"encoding/json"
)

func unmarshal(data []byte, v interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

func UnmarshalClientAttach(payload []byte) (string, error) {
	var str string
	err := unmarshal(payload, &str)
	return str, err
}

func UnmarshalClientGet(payload []byte) (string, error) {
	var str string
	err := unmarshal(payload, &str)
	return str, err
}

type Hook struct {
	Kind   string `json:"type"`
	Target string `json:"target"`
	Event  bool   `json:"event"`
}

// UnmarshalClientHook unmarshals the payload of the C_Hook message.
func UnmarshalClientHook(payload []byte) (*Hook, error) {
	var hook Hook
	err := unmarshal(payload, &hook)
	return &hook, err
}

// UnmarshalClientUnhook unmarshals the payload of the C_Unhook message.
func UnmarshalClientUnhook(payload []byte) (string, error) {
	var str string
	err := unmarshal(payload, &str)
	return str, err
}

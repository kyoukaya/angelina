package msg

import "encoding/json"

func UnmarshalClientAttach(payload []byte) (string, error) {
	var str string
	err := json.Unmarshal(payload, &str)
	return str, err
}

func UnmarshalClientGet(payload []byte) (string, error) {
	var str string
	err := json.Unmarshal(payload, &str)
	return str, err
}

type Hook struct {
	Kind   string `json:"type"`
	Target string `json:"target"`
	Event  bool   `json:"event,omitempty"`
}

// UnmarshalClientHook unmarshals the payload of C_Hook and C_Unhook messages.
func UnmarshalClientHook(payload []byte) (*Hook, error) {
	var hook Hook
	err := json.Unmarshal(payload, &hook)
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

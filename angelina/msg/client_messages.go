package msg

import "encoding/json"

type Hook struct {
	Kind   string `json:"type"`
	Target string `json:"target"`
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

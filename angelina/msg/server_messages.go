package msg

import (
	"encoding/json"
	"strconv"
)

func newBytes(old []byte) []byte {
	ret := make([]byte, len(old))
	copy(ret, old)
	return ret
}

var userList = []byte("S_UserList ")

// ServerUserList creates a message informing the client of the already attached
// users that are available to attach to.
func ServerUserList(users []string) ([]byte, error) {
	ret := newBytes(userList)
	b, err := json.Marshal(users)
	if err != nil {
		return nil, err
	}
	ret = append(ret, b...)
	return ret, nil
}

var serverNewUser = []byte("S_NewUser ")

// ServerNewUser creates a message notifying the client that a new user has
// connected through Rhine and is available to attach to.
func ServerNewUser(user string) ([]byte, error) {
	ret := newBytes(serverNewUser)
	b, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	ret = append(ret, b...)
	return ret, nil
}

var serverAttached = []byte("S_Attached ")

// ServerAttached creates a message notifying the client that they've successfully
// attached to a user.
func ServerAttached(user string) ([]byte, error) {
	ret := newBytes(serverAttached)
	b, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	ret = append(ret, b...)
	return ret, nil
}

var serverDetach = []byte("S_Detached")

// ServerDetach creates a message notifying the client that the user they were
// attached to has disconnected.
func ServerDetach() ([]byte, error) {
	return serverDetach, nil
}

var serverHooked = []byte("S_Hooked ")

type serverHookedT struct {
	ID     string `json:"id"`
	Kind   string `json:"type"`
	Target string `json:"target"`
	Event  bool   `json:"event,omitempty"`
}

// ServerHooked creates a message to notify the client that they have successfully
// registered a hook for an event.
func ServerHooked(id uint64, kind, target string, event bool) ([]byte, error) {
	ret := newBytes(serverHooked)
	res, err := json.Marshal(serverHookedT{
		ID:     strconv.FormatUint(id, 10),
		Kind:   kind,
		Target: target,
		Event:  event,
	})
	if err != nil {
		return nil, err
	}
	ret = append(ret, res...)
	return ret, nil
}

var serverUnhooked = []byte("S_Unhooked ")

type serverUnhookedT struct {
	ID     string `json:"id"`
	Kind   string `json:"type"`
	Target string `json:"target"`
	Event  bool   `json:"event,omitempty"`
}

// ServerUnhooked creates a message to notify that they have successfully unhooked
// for an event.
func ServerUnhooked(kind, target string) ([]byte, error) {
	ret := newBytes(serverUnhooked)
	res, err := json.Marshal(serverUnhookedT{
		Kind:   kind,
		Target: target,
	})
	if err != nil {
		return nil, err
	}
	ret = append(ret, res...)
	return ret, nil
}

var serverHookEvt = []byte("S_HookEvt ")

type serverHookEvtT struct {
	ID     string      `json:"id"`
	Kind   string      `json:"type"`
	Target string      `json:"target"`
	Data   interface{} `json:"data,omitempty"`
}

// ServerHookEvt notifies the client when a hook generates an event.
func ServerHookEvt(kind, target string, data interface{}) ([]byte, error) {
	ret := newBytes(serverHookEvt)
	res, err := json.Marshal(serverHookEvtT{
		Kind:   kind,
		Target: target,
		Data:   data,
	})
	if err != nil {
		return nil, err
	}
	ret = append(ret, res...)
	return ret, nil
}

var serverError = []byte("S_Error ")

type serverErrorT struct {
	Error   string `json:"error"`
	Request string `json:"request"`
}

// ServerError creates a message to notify the client that an error has occurred
// during the handling of a request.
func ServerError(request []byte, err string) ([]byte, error) {
	ret := newBytes(serverError)
	res, mErr := json.Marshal(serverErrorT{
		Error:   err,
		Request: string(request),
	})
	if mErr != nil {
		return nil, mErr
	}
	ret = append(ret, res...)
	return ret, nil
}

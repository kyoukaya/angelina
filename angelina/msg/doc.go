/*
Package msg provides marshalling and unmarshalling functions for messages over
websockets. Messages are formatted as such:
	{OP_CODE} {PAYLOAD}
With a single space delimiting the op code from the payload. Where the payload
is a json value, i.e., a json string/array/number/object, but may be omitted
if no payload is necessary.

Messages from the server to the client:
S_UserList - Sent on first connection with Angelina
	["string"]  // Array of user identifiers '{REGION}_{UID}'
S_NewUser - When a new user logs in through Rhine
	"string"  // User identifier '{REGION}_{UID}'
S_Attached
	"string"  // User identifier '{REGION}_{UID}'
S_Detached - When the connected user is disconnected
	No payload.
S_Hooked - On successful hook request.
	{
		"id": "string",  // Required for unhooking
		"type": "string",  // 'gamestate' or 'packet'
		"target": "string",
		"event": "boolean"  // Optional, false if not sent
	}
S_Unhooked - On successful unhook request.
	"string"
S_HookEvt - Sent when a hook generates an event.
	{
		"type": "string",  // 'gamestate' or 'packet'
		"target": "string",
		// data's JSON type may vary depending on the hook target.
		// Omitted if the hook is an event type
		"data": "data object"
	}
S_Get - Sent after the client sends a C_Get request if the get is successful.
	{
		"path": "string",
		"data": "data object"
	}
S_Error - Sent when an error was generated while handling of a request.
	{
		"request": "string",  // The request message that generated the error
		"error": "string"
	}

Messages from the client to the server:
C_Attach - C_Attach is sent from the websocket client to request for the server to
attach them to the specified game user. A websocket client can only be attached to
one user at a time and it is required for hooking and getting information from their game state.
	"string"
C_Detach - unhooks all registered hooks and allows the websocket client to attach to another user.
	No payload
C_Get - requests a piece of information from the attached user's game state.
	"string"
C_Hook - requests a hook to be made on either a certain packet being received or if there's
a change to the gamestate in a certain path. The event value specifies if the websocket
client only needs to be notified of the change or packet and not sent the data itself.
	{
		"type": "string",  // 'gamestate' or 'packet'
		"target": "string",
		"event": "boolean"  // Optional, defaults to false
	}
C_Unhook - stop listening on an event.
	"string"  // Hook ID
*/
package msg

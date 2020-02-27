# angelina

A websocket server built on top of [Rhine](https://github.com/kyoukaya/rhine) that acts as a messenger, allowing developers to listen on game events and pull information from a user's game state from any language with a websocket client.

## Protocol

Angelina communicates over text mode websockets, marshalling native data structures to JSON.
All messages begin with an opcode, if the message contains a payload it will be separated from the opcode with a single space. Messages received from the websocket client are guaranteed to be read and processed in the order that they were sent.
Example of communications with angelina over websockets, with the [websocat client](https://github.com/vi/websocat), can be seen below, comment lines prefaced with `//`.
Further documentation of individual messages can be found [here](https://github.com/kyoukaya/angelina/blob/master/angelina/msg/doc.go).

```
$ websocat ws://localhost:8000/ws
// Upon connecting with the angelina, the server sends S_UserList which contains an array of
// IDs of all game users already connected. The ID will be used for attaching to one of them.
S_UserList ["GL_99999"]
// If a game user connects after the websocket client connects, a S_NewUser message is sent
// to the connected websocket clients.
S_NewUser "JP_99999"
// Client messages always begin with "C_" while server messages begin with "S_"
// C_Attach is sent from the websocket client to request for the server to attach them to the
// specified game user. A websocket client can only be attached to one user at a time and it
// is required for hooking and getting information from their game state.
C_Attach "GL_99999"
// S_Attached is a confirmation from the server that the websocket client is attached.
S_Attached "GL_99999"
// Once attached, the websocket client can send C_Get, C_Hook, C_Detach messages.
// C_Hook requests a hook to be made on either a certain packet being received or if there's
// a change to the gamestate in a certain path. The event value specifies if the websocket
// client only needs to be notified of the change or packet and not sent the data itself.
C_Hook {"type":"gamestate", "target": "inventory", "event": false}
S_Hooked {"id":"0","type":"gamestate","target":"inventory","event":false}
// S_HookEvt are sent whenever the game user generates an event that triggers one of the hooks
// the websocket client has registered. The data field may be omitted if the hook is an event hook.
S_HookEvt {"type":"gamestate","target":"inventory","data":{"2001":271,"2002":41,"2003":25}}
// Clients can stop listening on an event with C_Unhook.
C_Unhook "0"
S_Unhooked "0"
// C_Get requests a piece of information from the attached user's game state.
C_Get "user"
// If an error occured during processing of any messages, the server will send a S_Error
// message containing the error and the message that caused the error.
S_Error {"error":"Unable to find the key","request":"C_Get \"user\""}
C_Get "status.socialPoint"
// S_Get returns the path of the request and the information requested.
S_Get {"path":"status.socialPoint","data":78}
// C_Detach unhooks all registered hooks and allows the websocket client to attach to another user.
C_Detach
S_Detached
```

## Examples

Example python3 client: [Recruitment Tag Calculator](https://github.com/kyoukaya/angelina/tree/dev/example_clients/python) prints recruitment tag combinations that guarantee a 4* or higher when recruiting.

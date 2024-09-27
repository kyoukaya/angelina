# angelina

A websocket server built on top of [Rhine](https://github.com/kyoukaya/rhine) that acts as a messenger, allowing clients to listen on game events and pull information from a user's game state from any programming language with a websocket library.

## Usage

angelina can be used as a rhine module and incorporated into a go program, see [main.go](https://github.com/kyoukaya/angelina/blob/master/cmd/main.go).
It can also be used standalone either from a [binary release](https://github.com/kyoukaya/angelina/releases) for users, or simply `go run cmd/main.go` for developers.

```
$ ./main.exe -help
Usage of C:\Users\kaya\Documents\ange\angelina\main.exe:
  -ange-host string
        host on which ange is served (default ":8000")
  -ange-static string
        path to static files to serve on the root URL. Serving disabled if empty string.
  -disable-cert-store
        disables the built in certstore, reduces memory usage but increases HTTP latency and CPU usage
  -filter
        enable the host filter
  -host string
        host on which the proxy is served (default ":8080")
  -log-path string
        file to output the log to (default "logs/proxy.log")
  -no-unk-json
        disallows unknown fields when unmarshalling json in the gamestate module
  -silent
        don't print anything to stdout
  -unsafe-origin
        allow any HTTP request, no matter what origin they specify, to upgrade into a ws connection
  -v    print Rhine verbose messages
  -v-goproxy
        print verbose goproxy messages
```

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

- [Recruitment Tag Calculator](https://github.com/kyoukaya/angelina/tree/master/example_clients/python) is a simple Python3 cli application that prints recruitment tag combinations that guarantee a 4* or higher when recruiting.
- [ifrit](https://github.com/kyoukaya/ifrit) is a multi-purpose web based toolkit written in Vue.js + Typescript currently in development.
angelina's static file serving capability is intended for use with such frameworks.

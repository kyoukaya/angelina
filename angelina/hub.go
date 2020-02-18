// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package angelina

import (
	"bytes"

	"github.com/kyoukaya/angelina/angelina/msg"
)

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			users := make([]string, 0, len(h.modules))
			for k := range h.modules {
				users = append(users, k)
			}
			res, err := msg.ServerUserList(users)
			if err != nil {
				h.Warnln(err)
				continue
			}
			client.sendWrapper(res)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case msg := <-h.messages:
			h.dispatch(msg)
		case mod := <-h.modAttach:
			userID := getModIdentifier(mod)
			h.modules[userID] = mod
			res, err := msg.ServerNewUser(userID)
			if err != nil {
				h.Warnln(err)
				continue
			}
			for client := range h.clients {
				client.sendWrapper(res)
			}
		case mod := <-h.modDetach:
			userID := getModIdentifier(mod)
			delete(h.modules, userID)
			res, err := msg.ServerDetach()
			if err != nil {
				h.Warnln(err)
				continue
			}
			for _, user := range h.users[userID] {
				user.user = ""
				user.sendWrapper(res)
			}
		}
	}
}

var spaceDemliter = []byte(" ")

func (h *Hub) dispatch(msg *message) {
	h.Printf("received message from %p:%s", msg.client, msg.payload)
	s := bytes.SplitN(msg.payload, spaceDemliter, 2)
	if len(s) == 1 {
		s = append(s, nil)
	}
	op := s[0]
	// payload := s[1]
	switch string(op) {
	case "C_Attach":
	case "C_Detach":
	case "C_Get":
	case "C_Hook":
	case "C_Unhook":
	}
}

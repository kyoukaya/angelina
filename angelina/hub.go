// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package angelina

import (
	"bytes"
	"fmt"

	"github.com/kyoukaya/angelina/angelina/msg"
)

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			// Build and send S_UserList
			users := make([]string, 0, len(h.modules))
			for k := range h.modules {
				users = append(users, k)
			}
			res, err := msg.ServerUserList(users)
			if err != nil {
				h.Warnln(err)
				h.sendErrorWrapper(client, err, []byte("register"))
				continue
			}
			client.sendWrapper(res)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.detachClient(client)
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
			for _, user := range h.attachedClients[userID] {
				user.user = ""
				user.sendWrapper(res)
			}
		}
	}
}

var spaceDemliter = []byte(" ")

func (h *Hub) dispatch(m *messageT) {
	h.Printf("received message from %p:%s", m.client, m.payload)
	s := bytes.SplitN(m.payload, spaceDemliter, 2)
	if len(s) == 1 {
		s = append(s, nil)
	}
	op := s[0]
	payload := s[1]
	// Get handler
	handler, exists := clientHandlerMap[string(op)]
	if !exists {
		err := fmt.Errorf("Unknown opcode '%s' received", op)
		h.Warnln(err)
		h.sendErrorWrapper(m.client, err, m.payload)
		return
	}
	err := handler(h, m.client, payload)
	if err != nil {
		h.sendErrorWrapper(m.client, err, m.payload)
	}
}

func (h *Hub) sendErrorWrapper(c *Client, err error, message []byte) {
	b, err := msg.ServerError(message, err.Error())
	if err != nil {
		h.Warnln(err)
		return
	}
	c.sendWrapper(b)
}

func (h *Hub) detachClient(client *Client) {
	var newClients []*Client
	for _, c := range h.attachedClients[client.user] {
		if c == client {
			continue
		}
		newClients = append(newClients, c)
	}
	h.attachedClients[client.user] = newClients
	client.user = ""
	// TODO: Unhook client's hooks
}

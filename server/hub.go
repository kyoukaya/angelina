package server

import (
	"bytes"
	"fmt"

	"github.com/kyoukaya/angelina/server/msg"
)

// Main event loop
// Synchronously handle client/module attaches, detaches, and messages.
func (ange *Ange) runHub() {
	for {
		select {
		// Handle new ws client connections
		case client := <-ange.register:
			ange.clients[client] = true
			// Build and send S_UserList
			users := make([]string, 0, len(ange.modules))
			for k := range ange.modules {
				users = append(users, k)
			}
			res, err := msg.ServerUserList(users)
			if err != nil {
				ange.Warnln("[Ange] ", err)
				ange.sendErrorWrapper(client, err, []byte("register"))
				continue
			}
			ange.Printf("[Ange] new websocket client %p", client)
			client.sendWrapper(res)
		// Handle ws client disconnects
		case client := <-ange.unregister:
			if _, ok := ange.clients[client]; ok {
				if client.mod != nil {
					ange.detachClient(client)
				}
				delete(ange.clients, client)
				close(client.send)
				ange.Printf("[Ange] websocket client disconnected %p", client)
			}
		// Handle messages from ws clients
		case msg := <-ange.messages:
			ange.dispatch(msg)
		// Handle new RhineModule connection
		case mod := <-ange.modAttach:
			userID := getModIdentifier(mod.RhineModule)
			ange.modules[userID] = mod
			res, err := msg.ServerNewUser(userID)
			if err != nil {
				ange.Warnln("[Ange] ", err)
				continue
			}
			for client := range ange.clients {
				client.sendWrapper(res)
			}
		// Handle new RhineModule disconnects
		case mod := <-ange.modDetach:
			userID := getModIdentifier(mod.RhineModule)
			delete(ange.modules, userID)
			detachMsg, err := msg.ServerDetach()
			if err != nil {
				ange.Warnln("[Ange] ", err)
				continue
			}
			for _, user := range ange.attachedClients[userID] {
				user.mod = nil
				user.sendWrapper(detachMsg)
			}
			ange.attachedClients[userID] = nil
		}
	}
}

var spaceDemliter = []byte(" ")

// Dispatch a client message to an appropriate handler.
func (ange *Ange) dispatch(m *messageT) {
	ange.Verbosef("[Ange] received message from %p:%s", m.client, m.payload)
	s := bytes.SplitN(m.payload, spaceDemliter, 2)
	if len(s) == 1 {
		s = append(s, nil)
	}
	op := s[0]
	payload := s[1]
	// Get handler
	handler, exists := clientHandlerMap[string(op)]
	if !exists {
		err := fmt.Errorf("[Ange] Unknown opcode '%s' received", op)
		ange.Warnln("[Ange] ", err)
		ange.sendErrorWrapper(m.client, err, m.payload)
		return
	}
	err := handler(ange, m.client, payload)
	if err != nil {
		ange.sendErrorWrapper(m.client, err, m.payload)
	}
}

func (ange *Ange) sendErrorWrapper(c *Client, err error, message []byte) {
	b, err := msg.ServerError(message, err.Error())
	if err != nil {
		ange.Warnln("[Ange] ", err)
		return
	}
	c.sendWrapper(b)
}

// detachClient detaches a client from a user by updating book keeping in the Hub
// and calling unhook on all their hooks. Calling this on a client that is not
// attached will result in a panic.
func (ange *Ange) detachClient(client *Client) {
	id := getModIdentifier(client.mod)
	clients := ange.attachedClients[id]
	i := 0
	for _, c := range clients {
		if c == client {
			break
		}
		i++
	}
	ange.attachedClients[id] = append(clients[:i], clients[i+1:]...)
	client.mod = nil
	client.unhookAll()
	ange.Printf("[Ange] detached %p from %s", client, id)
}

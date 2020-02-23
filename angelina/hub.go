package angelina

import (
	"bytes"
	"fmt"

	"github.com/kyoukaya/angelina/angelina/msg"
)

// Main event loop
func (h *Hub) run() {
	for {
		select {
		// Handle new ws client connections
		case client := <-h.register:
			h.clients[client] = true
			// Build and send S_UserList
			users := make([]string, 0, len(h.modules))
			for k := range h.modules {
				users = append(users, k)
			}
			res, err := msg.ServerUserList(users)
			if err != nil {
				h.Warnln("[Ange] ", err)
				h.sendErrorWrapper(client, err, []byte("register"))
				continue
			}
			h.Printf("[Ange] new websocket client %p", client)
			client.sendWrapper(res)
		// Handle ws client disconnects
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				if client.mod != nil {
					h.detachClient(client)
				}
				delete(h.clients, client)
				close(client.send)
				h.Printf("[Ange] websocket client disconnected %p", client)
			}
		// Handle messages from ws clients
		case msg := <-h.messages:
			h.dispatch(msg)
		// Handle new RhineModule connection
		case mod := <-h.modAttach:
			userID := getModIdentifier(mod.RhineModule)
			h.modules[userID] = mod
			res, err := msg.ServerNewUser(userID)
			if err != nil {
				h.Warnln("[Ange] ", err)
				continue
			}
			for client := range h.clients {
				client.sendWrapper(res)
			}
		// Handle new RhineModule disconnects
		case mod := <-h.modDetach:
			userID := getModIdentifier(mod.RhineModule)
			delete(h.modules, userID)
			detachMsg, err := msg.ServerDetach()
			if err != nil {
				h.Warnln("[Ange] ", err)
				continue
			}
			for _, user := range h.attachedClients[userID] {
				user.mod = nil
				user.sendWrapper(detachMsg)
			}
			h.attachedClients[userID] = nil
		}
	}
}

var spaceDemliter = []byte(" ")

func (h *Hub) dispatch(m *messageT) {
	h.Verbosef("[Ange] received message from %p:%s", m.client, m.payload)
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
		h.Warnln("[Ange] ", err)
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
		h.Warnln("[Ange] ", err)
		return
	}
	c.sendWrapper(b)
}

// detachClient detaches a client from a user by updating book keeping in the Hub
// and calling unhook on all their hooks. Calling this on a client that is not
// attached will result in a panic.
func (h *Hub) detachClient(client *Client) {
	id := getModIdentifier(client.mod)
	clients := h.attachedClients[id]
	i := 0
	for _, c := range clients {
		if c == client {
			break
		}
		i++
	}
	h.attachedClients[id] = append(clients[:i], clients[i+1:]...)
	client.mod = nil
	client.unhookAll()
	h.Printf("[Ange] detached %p from %s", client, id)
}

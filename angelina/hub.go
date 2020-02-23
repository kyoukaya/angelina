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
			client.sendWrapper(res)
		// Handle ws client disconnects
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.cleanupClient(client)
				delete(h.clients, client)
				close(client.send)
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

func (h *Hub) cleanupClient(client *Client) {
	var newClients []*Client
	// If the client is not attached, there's no state to clean up.
	if client.mod == nil {
		return
	}
	id := getModIdentifier(client.mod)
	for _, c := range h.attachedClients[id] {
		if c == client {
			continue
		}
		newClients = append(newClients, c)
	}
	h.attachedClients[id] = newClients
	client.mod = nil
	client.unhookAll()
}

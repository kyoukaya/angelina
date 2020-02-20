package angelina

import (
	"fmt"

	"github.com/kyoukaya/angelina/angelina/msg"
)

type clientMessageHandler func(h *Hub, client *Client, payload []byte) error

var clientHandlerMap = map[string]clientMessageHandler{
	"C_Attach": handleCAttach,
	"C_Detach": handleCDetach,
	"C_Get":    handleCGet,
	"C_Hook":   handleCHook,
	"C_Unhook": handleCUnhook,
}

func handleCAttach(h *Hub, client *Client, payload []byte) error {
	id, err := msg.UnmarshalClientAttach(payload)
	if err != nil {
		return err
	}

	if client.user != "" {
		return fmt.Errorf("Client is already connected to user '%s'", client.user)
	}

	if _, exists := h.modules[id]; !exists {
		return fmt.Errorf("User '%s' is not connected", id)
	}

	client.user = id
	h.attachedClients[id] = append(h.attachedClients[id], client)

	ret, err := msg.ServerAttached(id)
	if err != nil {
		return err
	}
	client.sendWrapper(ret)
	return nil
}

func handleCDetach(h *Hub, client *Client, payload []byte) error {
	if client.user == "" {
		return fmt.Errorf("User was not attached")
	}
	h.detachClient(client)

	ret, err := msg.ServerDetach()
	if err != nil {
		return err
	}
	client.sendWrapper(ret)
	return nil
}

func handleCGet(h *Hub, client *Client, payload []byte) error {
	return nil
}

func handleCHook(h *Hub, client *Client, payload []byte) error {
	return nil
}

func handleCUnhook(h *Hub, client *Client, payload []byte) error {
	return nil
}

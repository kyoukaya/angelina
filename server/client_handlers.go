package server

import (
	"fmt"
	"strconv"

	"github.com/kyoukaya/angelina/server/msg"
)

type clientMessageHandler func(h *Ange, client *Client, payload []byte) error

var clientHandlerMap = map[string]clientMessageHandler{
	"C_Attach": handleCAttach,
	"C_Detach": handleCDetach,
	"C_Get":    handleCGet,
	"C_Hook":   handleCHook,
	"C_Unhook": handleCUnhook,
}

func handleCAttach(h *Ange, client *Client, payload []byte) error {
	id, err := msg.UnmarshalClientAttach(payload)
	if err != nil {
		return err
	}

	if client.mod != nil {
		return fmt.Errorf("Client is already connected to user '%s'", getModIdentifier(client.mod))
	}

	mod, exists := h.modules[id]
	if !exists {
		return fmt.Errorf("User '%s' is not connected", id)
	}

	client.mod = mod.RhineModule
	h.attachedClients[id] = append(h.attachedClients[id], client)

	ret, err := msg.ServerAttached(id)
	if err != nil {
		return err
	}
	client.sendWrapper(ret)
	return nil
}

func handleCDetach(h *Ange, client *Client, payload []byte) error {
	if client.mod == nil {
		return fmt.Errorf("Client was not attached")
	}
	h.detachClient(client)
	ret, err := msg.ServerDetach()
	if err != nil {
		return err
	}
	client.sendWrapper(ret)
	return nil
}

func handleCGet(h *Ange, client *Client, payload []byte) error {
	if client.mod == nil {
		return fmt.Errorf("Client is not attached")
	}
	path, err := msg.UnmarshalClientGet(payload)
	if err != nil {
		return err
	}
	val, err := client.mod.StateGet(path)
	if err != nil {
		return err
	}
	ret, err := msg.ServerGet(path, val)
	if err != nil {
		return err
	}
	client.sendWrapper(ret)
	return nil
}

func handleCHook(h *Ange, client *Client, payload []byte) error {
	if client.mod == nil {
		return fmt.Errorf("Client is not attached")
	}
	data, err := msg.UnmarshalClientHook(payload)
	if err != nil {
		return err
	}
	return client.addHook(data)
}

func handleCUnhook(h *Ange, client *Client, payload []byte) error {
	idStr, err := msg.UnmarshalClientUnhook(payload)
	if err != nil {
		return err
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return err
	}
	err = client.removeHook(id)
	if err != nil {
		return err
	}
	ret, err := msg.ServerUnhooked(id)
	if err != nil {
		return err
	}
	client.sendWrapper(ret)
	return nil
}

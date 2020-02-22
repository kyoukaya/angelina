package angelina

import (
	"strconv"

	"github.com/kyoukaya/rhine/proxy"
)

const modName = "Angelina Module"

// angeModule is a Rhine module that serves as an interface for angelina to
// create and destroy hooks.
type angeModule struct {
	*proxy.RhineModule
	*Hub
}

func getModIdentifier(mod *proxy.RhineModule) string {
	return mod.Region + "_" + strconv.Itoa(mod.UID)
}

func (c *angeModule) shutdown(bool) {
	c.Hub.modDetach <- c
}

func (hub *Hub) modInitFunc(mod *proxy.RhineModule) {
	module := &angeModule{mod, hub}
	hub.modAttach <- module
	mod.OnShutdown(module.shutdown)
}

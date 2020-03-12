package server

import (
	"strconv"

	"github.com/kyoukaya/rhine/proxy"
)

const modName = "Angelina Module"

// angeModule is a Rhine module that serves as an interface for angelina to
// create and destroy hooks.
type angeModule struct {
	*proxy.RhineModule
	*Ange
}

func getModIdentifier(mod *proxy.RhineModule) string {
	return mod.Region + "_" + strconv.Itoa(mod.UID)
}

func (mod *angeModule) shutdown(bool) {
	mod.Ange.modDetach <- mod
}

func (hub *Ange) modInitFunc(mod *proxy.RhineModule) {
	module := &angeModule{mod, hub}
	hub.modAttach <- module
	mod.OnShutdown(module.shutdown)
}
